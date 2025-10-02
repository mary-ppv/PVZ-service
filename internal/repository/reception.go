package repository

import (
	"PVZ/internal/models"
	"PVZ/pkg/uuid"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"time"
)

type ReceptionRepo struct {
	db *sql.DB
}

func NewReceptionRepo(db *sql.DB) *ReceptionRepo {
	return &ReceptionRepo{db: db}
}

func (r *ReceptionRepo) CreateReception(pvzID string) (*models.Reception, error) {
	id, err := uuid.GenerateUUID7()
	if err != nil {
		log.Fatalf("Failed to generate UUIDv7: %v", err)
	}
	dateTime := time.Now()

	_, err = r.db.Exec(`
		INSERT INTO receptions (id, pvz_id, status, product_ids, date_time)
		VALUES ($1, $2, $3, $4, $5)
	`, id, pvzID, models.ReceptionInProgress, "[]", dateTime)
	if err != nil {
		return nil, err
	}

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
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	rec.PvzID = pvzID
	if err := json.Unmarshal([]byte(productIDsJSON), &rec.ProductIDs); err != nil {
		return nil, err
	}

	return &rec, nil
}

func (r *ReceptionRepo) CloseReception(receptionID string) error {
	_, err := r.db.Exec(`
		UPDATE receptions
		SET status = $1
		WHERE id = $2
	`, models.ReceptionClosed, receptionID)
	return err
}

func (r *ReceptionRepo) UpdateProducts(receptionID string, productIDs []string) error {
	data, _ := json.Marshal(productIDs)
	_, err := r.db.Exec(`
		UPDATE receptions
		SET product_ids = $1
		WHERE id = $2
	`, string(data), receptionID)
	return err
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
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(productIDsJSON), &rec.ProductIDs); err != nil {
		return nil, err
	}

	return &rec, nil
}

func (r *ReceptionRepo) DeleteLastProduct(receptionID string) (*models.Reception, error) {
	rec, err := r.GetByID(receptionID)
	if err != nil {
		return nil, err
	}
	if rec == nil || len(rec.ProductIDs) == 0 {
		return nil, errors.New("no products to delete")
	}

	rec.ProductIDs = rec.ProductIDs[:len(rec.ProductIDs)-1] // удалить последний
	if err := r.UpdateProducts(receptionID, rec.ProductIDs); err != nil {
		return nil, err
	}

	return rec, nil
}
