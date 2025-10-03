package service

import (
	"PVZ/internal/models"
	"PVZ/pkg/logger"
	"PVZ/pkg/metrics"
	"errors"
)

type ProductService struct {
	productRepo   ProductRepository
	receptionRepo ReceptionRepository
}

func NewProductService(pRepo ProductRepository, rRepo ReceptionRepository) *ProductService {
	return &ProductService{
		productRepo:   pRepo,
		receptionRepo: rRepo,
	}
}

func (s *ProductService) AddProduct(pvzID, userRole string, productType models.ProductType) (*models.Product, error) {
	if userRole != "employee" {
		logger.Log.Printf("Access denied: userRole=%s tried to add product to PVZ %s", userRole, pvzID)
		return nil, errors.New("access denied")
	}

	reception, err := s.receptionRepo.GetActiveByPVZ(pvzID)
	if err != nil {
		logger.Log.Printf("Failed to get active reception for PVZ %s: %v", pvzID, err)
		return nil, err
	}
	if reception == nil {
		logger.Log.Printf("No active reception found for PVZ %s", pvzID)
		return nil, errors.New("no active reception found")
	}

	product, err := s.productRepo.AddProduct(reception.ID, productType)
	if err != nil {
		logger.Log.Printf("Failed to add product to reception %s: %v", reception.ID, err)
		return nil, err
	}

	logger.Log.Printf("Product %s of type %s added to reception %s", product.ID, productType, reception.ID)
	metrics.ProductAdded.Inc()
	return product, nil
}
