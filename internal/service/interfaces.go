package service

import "PVZ/internal/models"

type UserRepository interface {
	GetByEmail(email string) (*models.User, error)
	CreateUser(user *models.User) error
}

type ProductRepository interface {
	AddProduct(receptionID string, productType models.ProductType) (*models.Product, error)
}

type PVZRepository interface {
	CreatePVZ(name string, city models.City) (*models.PVZ, error)
	GetPVZList(offset, limit int, cityFilter *models.City) ([]*models.PVZ, error)
}

type ReceptionRepository interface {
	CreateReception(pvzID string) (*models.Reception, error)
	GetActiveByPVZ(pvzID string) (*models.Reception, error)
	CloseReception(pvzID string) error
	DeleteLastProduct(receptionID string) (*models.Reception, error)
}
