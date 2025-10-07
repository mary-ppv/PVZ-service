package repository

import (
	"PVZ/models"
	"context"
	"log/slog"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

type PVZRepo struct {
	db boil.ContextExecutor
}

func NewPVZRepo(db boil.ContextExecutor) *PVZRepo {
	return &PVZRepo{db: db}
}

func (r *PVZRepo) CreatePVZ(ctx context.Context, name string, city string) (*models.PVZ, error) {
	pvz := &models.PVZ{
		Name:      name,
		City:      city,
		CreatedAt: time.Now(),
	}

	if err := pvz.Insert(ctx, r.db, boil.Infer()); err != nil {
		slog.Error("Failed to insert PVZ %s: %v", name, err)
		return nil, err
	}

	return pvz, nil
}

func (r *PVZRepo) GetPVZList(ctx context.Context, offset, limit int, cityFilter string) ([]*models.PVZ, error) {
	mods := []qm.QueryMod{
		qm.Limit(limit),
		qm.Offset(offset),
		qm.OrderBy(models.PVZColumns.CreatedAt + " DESC"),
	}

	if cityFilter != "" {
		mods = append(mods, models.PVZWhere.City.EQ(cityFilter))
	}

	pvzList, err := models.PVZS(mods...).All(ctx, r.db)
	if err != nil {
		slog.Error("Failed to get PVZ list: %v", err)
		return nil, err
	}

	return pvzList, nil
}
