package repository

import (
	"PVZ/internal/constants"
	"PVZ/models"
	"PVZ/pkg/uuid"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
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
		return nil, errors.New("failed to generate UUIDv7 for reception")
	}

	rec := &models.Reception{
		ID:     id,
		PVZID:  pvzIDInt,
		Status: constants.ReceptionInProgress,
	}

	if err := rec.Insert(ctx, r.db, boil.Infer()); err != nil {
		slog.Error("Failed to create reception %s: %v", id, err)
		return nil, err
	}

	return rec, nil
}

func (r *ReceptionRepo) GetActiveByPVZ(ctx context.Context, pvzID string) (*models.Reception, error) {
	pvzIDInt, err := strconv.ParseInt(pvzID, 10, 64)
	rec, err := models.Receptions(
		models.ReceptionWhere.PVZID.EQ(pvzIDInt),
		models.ReceptionWhere.Status.EQ(constants.ReceptionInProgress),
		qm.OrderBy(models.ReceptionColumns.DateTime+" DESC"),
		qm.Limit(1),
	).One(ctx, r.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		slog.Info("No active reception found for PVZ %s", pvzID)
		return nil, err
	}
	return rec, nil
}

func (r *ReceptionRepo) CloseReception(ctx context.Context, receptionID string) error {
	rec, err := models.FindReception(ctx, r.db, receptionID)
	if err != nil {
		slog.Warn("Failed to find reception %s: %v", receptionID, err)
		return err
	}

	rec.Status = constants.ReceptionClosed
	_, err = rec.Update(ctx, r.db, boil.Whitelist(models.ReceptionColumns.Status))
	if err != nil {
		slog.Error("Failed to close reception %s: %v", receptionID, err)
		return err
	}

	return nil
}

func (r *ReceptionRepo) UpdateProducts(ctx context.Context, receptionID string, productIDs []string) error {
	rec, err := models.FindReception(ctx, r.db, receptionID)
	if err != nil {
		slog.Warn("Failed to find reception %s: %v", receptionID, err)
		return err
	}

	jsonData, err := json.Marshal(productIDs)
	if err != nil {
		slog.Error("Failed to marshal product IDs: %v", err)
		return err
	}

	rec.ProductIds = types.JSON(jsonData)
	_, err = rec.Update(ctx, r.db, boil.Whitelist(models.ReceptionColumns.ProductIds))
	if err != nil {
		slog.Error("Failed to update products for reception %s: %v", receptionID, err)
		return err
	}

	return nil
}

func (r *ReceptionRepo) GetByID(ctx context.Context, receptionID string) (*models.Reception, error) {
	rec, err := models.FindReception(ctx, r.db, receptionID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return rec, err
}

func (r *ReceptionRepo) DeleteLastProduct(ctx context.Context, receptionID string) (*models.Reception, error) {
	rec, err := r.GetByID(ctx, receptionID)
	if err != nil || rec == nil {
		slog.Error("Can not get reception by id %s: %v", receptionID, err)
		return nil, err
	}

	var productIDs []string
	if err := rec.ProductIds.Unmarshal(&productIDs); err != nil {
		slog.Error("Failed to unmarshal reception IDs: %v", err)
		return nil, err
	}

	if len(productIDs) == 0 {
		return nil, errors.New("No products to delete")
	}

	productIDs = productIDs[:len(productIDs)-1]

	jsonData, err := json.Marshal(productIDs)
	if err != nil {
		slog.Error("Failed to marshal reception IDs: %v", err)
		return nil, err
	}

	rec.ProductIds = types.JSON(jsonData)

	_, err = rec.Update(ctx, r.db, boil.Whitelist(models.ReceptionColumns.ProductIds))
	if err != nil {
		slog.Error("Failed to update reception %s: %v", receptionID, err)
		return nil, err
	}

	return rec, nil
}
