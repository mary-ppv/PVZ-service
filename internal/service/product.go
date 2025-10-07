package service

import (
	"PVZ/internal/constants"
	"PVZ/internal/repository"
	"PVZ/models"
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
	if userRole != constants.RoleEmployee {
		return nil, errors.New("access denied")
	}

	validTypes := map[string]bool{
		"электроника": true,
		"одежда":      true,
		"обувь":       true,
	}

	if !validTypes[productType] {
		return nil, errors.New("invalid product type")
	}

	reception, err := s.receptionRepo.GetActiveByPVZ(ctx, pvzID)
	if err != nil {
		return nil, errors.New("failed to get active reception")
	}

	if reception == nil {
		return nil, errors.New("no active reception found")
	}

	if reception.Status != constants.ReceptionInProgress {
		return nil, errors.New("reception is not active. Status")
	}

	product, err := s.productRepo.AddProduct(ctx, reception.ID, productType)
	if err != nil {
		return nil, errors.New("failed to add product to reception")
	}

	metrics.ProductAdded.Inc()
	return product, nil
}
