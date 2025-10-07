package service

import (
	"PVZ/internal/constants"
	"PVZ/models"
	"PVZ/pkg/metrics"
	"context"
	"errors"
	"log/slog"
)

type ReceptionService struct {
	repo ReceptionRepository
}

func NewReceptionService(repo ReceptionRepository) *ReceptionService {
	return &ReceptionService{repo: repo}
}

func (s *ReceptionService) CreateReception(ctx context.Context, pvzID, userRole string) (*models.Reception, error) {
	if userRole != constants.RoleEmployee {
		return nil, errors.New("Access denied")
	}

	active, err := s.repo.GetActiveByPVZ(ctx, pvzID)
	if err != nil {
		slog.Error("Failed to get active reception", err)
		return nil, err
	}
	if active != nil {
		return nil, errors.New("There is already an active reception")
	}

	rec, err := s.repo.CreateReception(ctx, pvzID)
	if err != nil {
		slog.Error("Failed to create reception", err)
		return nil, err
	}

	rec.Status = constants.ReceptionInProgress

	metrics.ReceptionCreated.Inc()
	return rec, nil
}

func (s *ReceptionService) CloseReception(ctx context.Context, pvzID, userRole string) (*models.Reception, error) {
	if userRole != constants.RoleEmployee {
		return nil, errors.New("Access denied")
	}

	active, err := s.repo.GetActiveByPVZ(ctx, pvzID)
	if err != nil {
		slog.Error("Failed to get active reception for PVZ %s: %v", pvzID, err)
		return nil, err
	}
	if active == nil {
		return nil, errors.New("no active reception to close")
	}

	if err := s.repo.CloseReception(ctx, active.ID); err != nil {
		slog.Error("Failed to close reception %s: %v", active.ID, err)
		return nil, err
	}

	active.Status = constants.ReceptionClosed
	return active, nil
}

func (s *ReceptionService) DeleteLastProduct(ctx context.Context, pvzID, userRole string) (*models.Reception, error) {
	if userRole != constants.RoleEmployee {
		return nil, errors.New("Access denied")
	}

	active, err := s.repo.GetActiveByPVZ(ctx, pvzID)
	if err != nil {
		slog.Error("Failed to get active reception for PVZ %s: %v", pvzID, err)
		return nil, err
	}
	if active == nil {
		return nil, errors.New("no active reception")
	}

	rec, err := s.repo.DeleteLastProduct(ctx, active.ID)
	if err != nil {
		slog.Error("Failed to delete last product for reception %s: %v", active.ID, err)
		return nil, err
	}

	return rec, nil
}
