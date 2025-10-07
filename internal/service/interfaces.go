package service

import (
	"PVZ/models"
	"context"
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
}

type ProductRepository interface {
	AddProduct(ctx context.Context, receptionID string, productType string) (*models.Product, error)
}

type PVZRepository interface {
	CreatePVZ(ctx context.Context, name string, city string) (*models.PVZ, error)
	GetPVZList(ctx context.Context, offset, limit int, city string) ([]*models.PVZ, error)
}

type ReceptionRepository interface {
	CreateReception(ctx context.Context, pvzID string) (*models.Reception, error)
	GetActiveByPVZ(ctx context.Context, pvzID string) (*models.Reception, error)
	CloseReception(ctx context.Context, pvzID string) error
	DeleteLastProduct(ctx context.Context, receptionID string) (*models.Reception, error)
}
