package repository

import (
	"context"
	"database/sql"

	"PVZ/internal/models"
)

type PVZRepository interface {
	Create(ctx context.Context, pvz *models.PVZ) (int64, error)
	GetByID(ctx context.Context, id int64) (*models.PVZ, error)
	List(ctx context.Context, startDate, endDate string, limit, offset int) ([]models.PVZ, error)
}

type pvzRepository struct {
	db *sql.DB
}

func NewPVZRepository(db *sql.DB) PVZRepository {
	return &pvzRepository{db: db}
}

func (r *pvzRepository) Create(ctx context.Context, pvz *models.PVZ) (int64, error) {
	query := `INSERT INTO pvz (name, city, created_at) VALUES ($1, $2, $3) RETURNING id`
	var id int64
	err := r.db.QueryRowContext(ctx, query, pvz.Name, pvz.City, pvz.CreatedAt).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *pvzRepository) GetByID(ctx context.Context, id int64) (*models.PVZ, error) {
	query := `SELECT id, name, city, created_at FROM pvz WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var p models.PVZ
	if err := row.Scan(&p.ID, &p.Name, &p.City, &p.CreatedAt); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *pvzRepository) List(ctx context.Context, startDate, endDate string, limit, offset int) ([]models.PVZ, error) {
	query := `
		SELECT id, name, city, created_at
		FROM pvz
		WHERE ($1 = '' OR created_at >= $1)
		  AND ($2 = '' OR created_at <= $2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, startDate, endDate, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pvzs []models.PVZ
	for rows.Next() {
		var p models.PVZ
		if err := rows.Scan(&p.ID, &p.Name, &p.City, &p.CreatedAt); err != nil {
			return nil, err
		}
		pvzs = append(pvzs, p)
	}
	return pvzs, nil

}
