package service

import (
	"PVZ/models"
	"PVZ/pkg/logger"
	"PVZ/pkg/metrics"
	"context"
	"errors"
)

type ReceptionService struct {
	repo ReceptionRepository
}

func NewReceptionService(repo ReceptionRepository) *ReceptionService {
	return &ReceptionService{repo: repo}
}

func (s *ReceptionService) CreateReception(ctx context.Context, pvzID, userRole string) (*models.Reception, error) {
	if userRole != "employee" {
		logger.Log.Printf("Access denied: userRole=%s tried to create reception for PVZ %s", userRole, pvzID)
		return nil, errors.New("access denied")
	}

	active, err := s.repo.GetActiveByPVZ(ctx, pvzID)
	if err != nil {
		logger.Log.Printf("Failed to get active reception for PVZ %s: %v", pvzID, err)
		return nil, err
	}
	if active != nil {
		logger.Log.Printf("Cannot create reception: active reception already exists for PVZ %s", pvzID)
		return nil, errors.New("there is already an active reception")
	}

	rec, err := s.repo.CreateReception(ctx, pvzID)
	if err != nil {
		logger.Log.Printf("Failed to create reception for PVZ %s: %v", pvzID, err)
		return nil, err
	}

	logger.Log.Printf("Reception %s created for PVZ %s", rec.ID, pvzID)
	metrics.ReceptionCreated.Inc()
	return rec, nil
}

func (s *ReceptionService) CloseReception(ctx context.Context, pvzID, userRole string) (*models.Reception, error) {
	if userRole != "employee" {
		logger.Log.Printf("Access denied: userRole=%s tried to close reception for PVZ %s", userRole, pvzID)
		return nil, errors.New("access denied")
	}

	active, err := s.repo.GetActiveByPVZ(ctx, pvzID)
	if err != nil {
		logger.Log.Printf("Failed to get active reception for PVZ %s: %v", pvzID, err)
		return nil, err
	}
	if active == nil {
		logger.Log.Printf("No active reception to close for PVZ %s", pvzID)
		return nil, errors.New("no active reception to close")
	}

	if err := s.repo.CloseReception(ctx, active.ID); err != nil {
		logger.Log.Printf("Failed to close reception %s: %v", active.ID, err)
		return nil, err
	}

	active.Status = models.ReceptionClosed
	logger.Log.Printf("Reception %s closed for PVZ %s", active.ID, pvzID)
	return active, nil
}

func (s *ReceptionService) DeleteLastProduct(ctx context.Context, pvzID, userRole string) (*models.Reception, error) {
	if userRole != "employee" {
		logger.Log.Printf("Access denied: userRole=%s tried to delete last product for PVZ %s", userRole, pvzID)
		return nil, errors.New("access denied")
	}

	active, err := s.repo.GetActiveByPVZ(ctx, pvzID)
	if err != nil {
		logger.Log.Printf("Failed to get active reception for PVZ %s: %v", pvzID, err)
		return nil, err
	}
	if active == nil {
		logger.Log.Printf("No active reception for PVZ %s to delete last product", pvzID)
		return nil, errors.New("no active reception")
	}

	rec, err := s.repo.DeleteLastProduct(ctx, active.ID)
	if err != nil {
		logger.Log.Printf("Failed to delete last product for reception %s: %v", active.ID, err)
		return nil, err
	}

	logger.Log.Printf("Deleted last product for reception %s, remaining products: %v", active.ID, rec.ProductIds)
	return rec, nil
}
