package repository

import (
	"PVZ/internal/models"
	"PVZ/pkg/database"
	"PVZ/pkg/uuid"
	"encoding/json"
	"log"
	"time"
)

type ProductRepo struct {
	db *database.DB
}

func NewProductRepo(db *database.DB) *ProductRepo {
	return &ProductRepo{db: db}
}

func (r *ProductRepo) AddProduct(receptionID string, productType models.ProductType) (*models.Product, error) {
	id, err := uuid.GenerateUUID7()
	if err != nil {
		log.Fatalf("Failed to generate UUIDv7: %v", err)
	}
	addedAt := time.Now()

	_, err = r.db.Exec(`
		INSERT INTO products (id, reception_id, type, added_at)
		VALUES ($1, $2, $3, $4)
	`, id, receptionID, productType, addedAt)
	if err != nil {
		return nil, err
	}

	var productIDsJSON string
	err = r.db.QueryRow(`SELECT product_ids FROM receptions WHERE id = $1`, receptionID).Scan(&productIDsJSON)
	if err != nil {
		return nil, err
	}

	var productIDs []string
	if err := json.Unmarshal([]byte(productIDsJSON), &productIDs); err != nil {
		return nil, err
	}

	productIDs = append(productIDs, id)
	updatedJSON, _ := json.Marshal(productIDs)

	_, err = r.db.Exec(`UPDATE receptions SET product_ids = $1 WHERE id = $2`, string(updatedJSON), receptionID)
	if err != nil {
		return nil, err
	}

	return &models.Product{
		ID:          id,
		ReceptionID: receptionID,
		Type:        productType,
		AddedAt:     addedAt,
	}, nil
}
