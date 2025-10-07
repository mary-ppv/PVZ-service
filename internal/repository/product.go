package repository

import (
	"PVZ/models"
	"PVZ/pkg/uuid"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/types"
)

type ProductRepo struct {
	db boil.ContextExecutor
}

func NewProductRepo(db boil.ContextExecutor) *ProductRepo {
	return &ProductRepo{db: db}
}

func (r *ProductRepo) AddProduct(ctx context.Context, receptionID string, productType string) (*models.Product, error) {
	id, err := uuid.GenerateUUID7()
	if err != nil {
		return nil, errors.New("Failed to generate UUIDv7")
	}

	product := &models.Product{
		ID:          id,
		ReceptionID: receptionID,
		Type:        productType,
		AddedAt:     time.Now(),
	}

	if err := product.Insert(ctx, r.db, boil.Infer()); err != nil {
		slog.Error("Failed to insert product: %v", err)
		return nil, err
	}

	rec, err := models.Receptions(models.ReceptionWhere.ID.EQ(receptionID)).One(ctx, r.db)
	if err != nil {
		return nil, errors.New("Failed to get reception")
	}

	var productIds []string
	if err := json.Unmarshal([]byte(rec.ProductIds), &productIds); err != nil {
		return nil, errors.New("Failed to unmarshal product IDs")
	}

	productIds = append(productIds, product.ID)
	updatedJSON, err := json.Marshal(productIds)
	if err != nil {
		slog.Error("Failed to marshal product IDs: %v", err)
		return nil, err
	}
	rec.ProductIds = types.JSON(updatedJSON)

	if _, err := rec.Update(ctx, r.db, boil.Whitelist("product_ids")); err != nil {
		return nil, errors.New("Failed to update reception product_ids")
	}

	return product, nil
}
