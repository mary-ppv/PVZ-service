package service

import (
	"PVZ/internal/repository"
	"PVZ/models"
	"PVZ/pkg/logger"
	"PVZ/pkg/metrics"
	"context"
	"errors"
)

type ProductService struct {
	productRepo   ProductRepository
	receptionRepo ReceptionRepository
}

func NewProductService(pRepo *repository.ProductRepo, rRepo *repository.ReceptionRepo) *ProductService {
	return &ProductService{
		productRepo:   pRepo,
		receptionRepo: rRepo,
	}
}

func (s *ProductService) AddProduct(ctx context.Context, pvzID, userRole string, productType string) (*models.Product, error) {
	if userRole != "employee" {
		logger.Log.Printf("Access denied: userRole=%s tried to add product to PVZ %s", userRole, pvzID)
		return nil, errors.New("access denied")
	}

	reception, err := s.receptionRepo.GetActiveByPVZ(ctx, pvzID)
	if err != nil {
		logger.Log.Printf("Failed to get active reception for PVZ %s: %v", pvzID, err)
		return nil, err
	}
	if reception == nil {
		logger.Log.Printf("No active reception found for PVZ %s", pvzID)
		return nil, errors.New("no active reception found")
	}

	product, err := s.productRepo.AddProduct(ctx, reception.ID, productType)
	if err != nil {
		logger.Log.Printf("Failed to add product to reception %s: %v", reception.ID, err)
		return nil, err
	}

	logger.Log.Printf("Product %s of type %s added to reception %s", product.ID, productType, reception.ID)
	metrics.ProductAdded.Inc()
	return product, nil
}
