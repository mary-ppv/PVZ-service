package service

import (
	"PVZ/internal/models"
	"PVZ/internal/repository"
	"context"
	"errors"
)

type ReceptionService interface {
	CreateReception(ctx context.Context, pvzID, userRole string) (*models.Reception, error)
	CloseReception(ctx context.Context, pvzID, userRole string) (*models.Reception, error)
}

type receptionService struct {
	repo repository.ReceptionRepository
}

func NewReceptionService(repo repository.ReceptionRepository) ReceptionService {
	return &receptionService{repo: repo}
}

func (s *receptionService) CreateReception(ctx context.Context, pvzID, userRole string) (*models.Reception, error) {
	if userRole != "employee" {
		return nil, errors.New("access denied")
	}

	active, err := s.repo.GetActiveByPVZ(pvzID)
	if err != nil {
		return nil, err
	}
	if active != nil {
		return nil, errors.New("there is already an active reception")
	}

	return s.repo.CreateReception(pvzID)
}

func (s *receptionService) CloseReception(ctx context.Context, pvzID, userRole string) (*models.Reception, error) {
	if userRole != "employee" {
		return nil, errors.New("access denied")
	}

	active, err := s.repo.GetActiveByPVZ(pvzID)
	if err != nil {
		return nil, err
	}
	if active == nil {
		return nil, errors.New("no active reception to close")
	}

	if err := s.repo.CloseReception(active.ID); err != nil {
		return nil, err
	}

	active.Status = models.ReceptionClosed
	return active, nil
}
