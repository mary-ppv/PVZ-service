package repository

import (
	"PVZ/internal/models"
	"database/sql"
	"encoding/json"
	"time"
)

type ReceptionRepository interface {
	CreateReception(pvzID string) (*models.Reception, error)
	GetActiveByPVZ(pvzID string) (*models.Reception, error)
	CloseReception(receptionID string) error
	UpdateProducts(receptionID string, productIDs []string) error
}

type receptionRepo struct {
	db *sql.DB
}

func NewReceptionRepository(db *sql.DB) ReceptionRepository {
	return &receptionRepo{db: db}
}

func (r *receptionRepo) CreateReception(pvzID string) (*models.Reception, error) {
	id := GenerateUUID()
	dateTime := time.Now()

	_, err := r.db.Exec(`
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

func (r *receptionRepo) GetActiveByPVZ(pvzID string) (*models.Reception, error) {
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

func (r *receptionRepo) CloseReception(receptionID string) error {
	_, err := r.db.Exec(`
		UPDATE receptions
		SET status = $1
		WHERE id = $2
	`, models.ReceptionClosed, receptionID)
	return err
}

func (r *receptionRepo) UpdateProducts(receptionID string, productIDs []string) error {
	data, _ := json.Marshal(productIDs)
	_, err := r.db.Exec(`
		UPDATE receptions
		SET product_ids = $1
		WHERE id = $2
	`, string(data), receptionID)
	return err
}
