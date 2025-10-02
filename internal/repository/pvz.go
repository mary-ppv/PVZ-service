package repository

import (
	"PVZ/internal/models"
	"database/sql"
	"time"
)

type PVZRepo struct {
	db *sql.DB
}

func NewPVZRepo(db *sql.DB) *PVZRepo {
	return &PVZRepo{db: db}
}

func (r *PVZRepo) CreatePVZ(name string, city models.City) (*models.PVZ, error) {
	createdAt := time.Now()
	res, err := r.db.Exec(`
		INSERT INTO pvz (name, city, created_at)
		VALUES (?, ?, ?)
	`, name, city, createdAt)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.PVZ{
		ID:        id,
		Name:      name,
		City:      city,
		CreatedAt: createdAt,
	}, nil
}

func (r *PVZRepo) GetPVZList(offset, limit int, cityFilter *models.City) ([]*models.PVZ, error) {
	query := `SELECT id, name, city, created_at FROM pvz`
	args := []interface{}{}

	if cityFilter != nil {
		query += " WHERE city = ?"
		args = append(args, *cityFilter)
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pvzList []*models.PVZ
	for rows.Next() {
		var pvz models.PVZ
		if err := rows.Scan(&pvz.ID, &pvz.Name, &pvz.City, &pvz.CreatedAt); err != nil {
			return nil, err
		}
		pvzList = append(pvzList, &pvz)
	}

	return pvzList, nil
}
