package service

import (
	"PVZ/internal/models"
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
		return nil, errors.New("access denied")
	}

	reception, err := s.receptionRepo.GetActiveByPVZ(pvzID)
	if err != nil {
		return nil, err
	}
	if reception == nil {
		return nil, errors.New("no active reception found")
	}

	return s.productRepo.AddProduct(reception.ID, productType)
}
