package repository

import (
	"PVZ/models"
	"PVZ/pkg/logger"
	"PVZ/pkg/uuid"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/aarondl/sqlboiler/v4/types"
)

type ReceptionRepo struct {
	db boil.ContextExecutor
}

func NewReceptionRepo(db boil.ContextExecutor) *ReceptionRepo {
	return &ReceptionRepo{db: db}
}

func (r *ReceptionRepo) CreateReception(ctx context.Context, pvzID string) (*models.Reception, error) {
	pvzIDInt, err := strconv.ParseInt(pvzID, 10, 64)
	if err != nil {
		return nil, errors.New("invalid PVZ ID format")
	}

	id, err := uuid.GenerateUUID7()
	if err != nil {
		logger.Log.Printf("Failed to generate UUIDv7 for reception: %v", err)
		return nil, err
	}

	rec := &models.Reception{
		ID:     id,
		PVZID:  pvzIDInt,
		Status: models.ReceptionInProgress,
	}

	if err := rec.Insert(ctx, r.db, boil.Infer()); err != nil {
		logger.Log.Printf("Failed to create reception %s: %v", id, err)
		return nil, err
	}

	logger.Log.Printf("Reception %s created for PVZ %d", id, pvzIDInt)
	return rec, nil
}

func (r *ReceptionRepo) GetActiveByPVZ(ctx context.Context, pvzID string) (*models.Reception, error) {
	pvzIDInt, err := strconv.ParseInt(pvzID, 10, 64)
	rec, err := models.Receptions(
		models.ReceptionWhere.PVZID.EQ(pvzIDInt),
		models.ReceptionWhere.Status.EQ(models.ReceptionInProgress),
		qm.OrderBy(models.ReceptionColumns.DateTime+" DESC"),
		qm.Limit(1),
	).One(ctx, r.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Log.Printf("No active reception found for PVZ %d", pvzIDInt)
			return nil, nil
		}
		logger.Log.Printf("Failed to get active reception for PVZ %d: %v", pvzID, err)
		return nil, err
	}
	return rec, nil
}

func (r *ReceptionRepo) CloseReception(ctx context.Context, receptionID string) error {
	rec, err := models.FindReception(ctx, r.db, receptionID)
	if err != nil {
		return err
	}

	rec.Status = models.ReceptionClosed
	_, err = rec.Update(ctx, r.db, boil.Whitelist(models.ReceptionColumns.Status))
	if err != nil {
		logger.Log.Printf("Failed to close reception %s: %v", receptionID, err)
		return err
	}

	logger.Log.Printf("Reception %s closed", receptionID)
	return nil
}

func (r *ReceptionRepo) UpdateProducts(ctx context.Context, receptionID string, productIDs []string) error {
	rec, err := models.FindReception(ctx, r.db, receptionID)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(productIDs)
	if err != nil {
		return err
	}

	rec.ProductIds = types.JSON(jsonData)
	_, err = rec.Update(ctx, r.db, boil.Whitelist(models.ReceptionColumns.ProductIds))
	if err != nil {
		logger.Log.Printf("Failed to update products for reception %s: %v", receptionID, err)
		return err
	}

	logger.Log.Printf("Updated products for reception %s: %v", receptionID, productIDs)
	return nil
}

func (r *ReceptionRepo) GetByID(ctx context.Context, receptionID string) (*models.Reception, error) {
	rec, err := models.FindReception(ctx, r.db, receptionID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return rec, err
}

func (r *ReceptionRepo) DeleteLastProduct(ctx context.Context, receptionID string) (*models.Reception, error) {
	rec, err := r.GetByID(ctx, receptionID)
	if err != nil || rec == nil {
		return nil, err
	}

	var productIDs []string
	if err := rec.ProductIds.Unmarshal(&productIDs); err != nil {
		return nil, err
	}

	if len(productIDs) == 0 {
		return nil, errors.New("no products to delete")
	}

	productIDs = productIDs[:len(productIDs)-1]

	jsonData, err := json.Marshal(productIDs)
	if err != nil {
		return nil, err
	}

	rec.ProductIds = types.JSON(jsonData)

	_, err = rec.Update(ctx, r.db, boil.Whitelist(models.ReceptionColumns.ProductIds))
	if err != nil {
		return nil, err
	}

	return rec, nil
}
