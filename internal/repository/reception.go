package repository

import (
	"PVZ/internal/models"
	"PVZ/pkg/database"
	"PVZ/pkg/logger"
	"PVZ/pkg/uuid"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type ReceptionRepo struct {
	db *database.DB
}

func NewReceptionRepo(db *database.DB) *ReceptionRepo {
	return &ReceptionRepo{db: db}
}

func (r *ReceptionRepo) CreateReception(pvzID string) (*models.Reception, error) {
	id, err := uuid.GenerateUUID7()
	if err != nil {
		logger.Log.Printf("Failed to generate UUIDv7 for reception: %v", err)
	}
	dateTime := time.Now()

	_, err = r.db.Exec(`
		INSERT INTO receptions (id, pvz_id, status, product_ids, date_time)
		VALUES ($1, $2, $3, $4, $5)
	`, id, pvzID, models.ReceptionInProgress, "[]", dateTime)
	if err != nil {
		logger.Log.Printf("Failed to create reception %s: %v", id, err)
		return nil, err
	}

	logger.Log.Printf("Reception %s created for PVZ %s", id, pvzID)

	return &models.Reception{
		ID:         id,
		PvzID:      pvzID,
		Status:     models.ReceptionInProgress,
		DateTime:   dateTime,
		ProductIDs: []string{},
	}, nil
}

func (r *ReceptionRepo) GetActiveByPVZ(pvzID string) (*models.Reception, error) {
	row := r.db.QueryRow(`
		SELECT id, status, product_ids, date_time
		FROM receptions
		WHERE pvz_id = $1 AND status = $2
		ORDER BY date_time DESC
		LIMIT 1
	`, pvzID, models.ReceptionInProgress)

	var rec models.Reception
	var productIDsJSON string
	err := row.Scan(&rec.ID, &rec.Status, &productIDsJSON, &rec.DateTime)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Log.Printf("No active reception found for PVZ %s", pvzID)
		return nil, nil
	}
	if err != nil {
		logger.Log.Printf("Failed to get active reception for PVZ %s: %v", pvzID, err)
		return nil, err
	}

	rec.PvzID = pvzID
	if err := json.Unmarshal([]byte(productIDsJSON), &rec.ProductIDs); err != nil {
		logger.Log.Printf("Failed to unmarshal product IDs for reception %s: %v", rec.ID, err)
		return nil, err
	}

	logger.Log.Printf("Active reception %s retrieved for PVZ %s", rec.ID, pvzID)

	return &rec, nil
}

func (r *ReceptionRepo) CloseReception(receptionID string) error {
	_, err := r.db.Exec(`
		UPDATE receptions
		SET status = $1
		WHERE id = $2
	`, models.ReceptionClosed, receptionID)

	if err != nil {
		logger.Log.Printf("Failed to close reception %s: %v", receptionID, err)
		return err
	}

	logger.Log.Printf("Reception %s closed", receptionID)
	return nil
}

func (r *ReceptionRepo) UpdateProducts(receptionID string, productIDs []string) error {
	data, _ := json.Marshal(productIDs)
	_, err := r.db.Exec(`
		UPDATE receptions
		SET product_ids = $1
		WHERE id = $2
	`, string(data), receptionID)

	if err != nil {
		logger.Log.Printf("Failed to update products for reception %s: %v", receptionID, err)
		return err
	}

	logger.Log.Printf("Updated products for reception %s: %v", receptionID, productIDs)
	return nil
}

func (r *ReceptionRepo) GetByID(receptionID string) (*models.Reception, error) {
	row := r.db.QueryRow(`
		SELECT id, pvz_id, status, product_ids, date_time
		FROM receptions
		WHERE id = $1
	`, receptionID)

	var rec models.Reception
	var productIDsJSON string
	err := row.Scan(&rec.ID, &rec.PvzID, &rec.Status, &productIDsJSON, &rec.DateTime)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Log.Printf("Reception %s not found", receptionID)
		return nil, nil
	}
	if err != nil {
		logger.Log.Printf("Failed to get reception %s: %v", receptionID, err)
		return nil, err
	}

	if err := json.Unmarshal([]byte(productIDsJSON), &rec.ProductIDs); err != nil {
		logger.Log.Printf("Failed to unmarshal product IDs for reception %s: %v", rec.ID, err)
		return nil, err
	}

	logger.Log.Printf("Reception %s retrieved", rec.ID)
	return &rec, nil
}

func (r *ReceptionRepo) DeleteLastProduct(receptionID string) (*models.Reception, error) {
	rec, err := r.GetByID(receptionID)
	if err != nil {
		logger.Log.Printf("Failed to get reception %s for deleting last product: %v", receptionID, err)
		return nil, err
	}
	if rec == nil || len(rec.ProductIDs) == 0 {
		logger.Log.Printf("No products to delete for reception %s", receptionID)
		return nil, errors.New("no products to delete")
	}

	rec.ProductIDs = rec.ProductIDs[:len(rec.ProductIDs)-1]
	if err := r.UpdateProducts(receptionID, rec.ProductIDs); err != nil {
		logger.Log.Printf("Failed to update products after deleting last product for reception %s: %v", receptionID, err)
		return nil, err
	}

	logger.Log.Printf("Deleted last product for reception %s, remaining products: %v", receptionID, rec.ProductIDs)
	return rec, nil
}
