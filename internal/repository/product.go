package repository

import (
	"PVZ/models"
	"PVZ/pkg/logger"
	"PVZ/pkg/uuid"
	"context"
	"encoding/json"
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
		logger.Log.Printf("Failed to generate UUIDv7: %v", err)
	}

	product := &models.Product{
		ID:          id,
		ReceptionID: receptionID,
		Type:        productType,
		AddedAt:     time.Now(),
	}

	if err := product.Insert(ctx, r.db, boil.Infer()); err != nil {
		logger.Log.Printf("Failed to insert product: %v", err)
		return nil, err
	}

	rec, err := models.Receptions(models.ReceptionWhere.ID.EQ(receptionID)).One(ctx, r.db)
	if err != nil {
		logger.Log.Printf("Failed to get reception %s: %v", receptionID, err)
		return nil, err
	}

	var productIds []string
	if err := json.Unmarshal([]byte(rec.ProductIds), &productIds); err != nil {
		logger.Log.Printf("Failed to unmarshal product IDs: %v", err)
		return nil, err
	}

	productIds = append(productIds, product.ID)
	updatedJSON, _ := json.Marshal(productIds)
	rec.ProductIds = types.JSON(updatedJSON)

	if _, err := rec.Update(ctx, r.db, boil.Whitelist("product_ids")); err != nil {
		logger.Log.Printf("Failed to update reception product_ids: %v", err)
		return nil, err
	}

	logger.Log.Printf("Product %s added to reception %s", product.ID, receptionID)
	return product, nil
}
