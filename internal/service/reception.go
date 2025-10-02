package service

import (
	"PVZ/internal/models"
	"errors"
)

type ReceptionService struct {
	repo ReceptionRepository
}

func NewReceptionService(repo ReceptionRepository) *ReceptionService {
	return &ReceptionService{repo: repo}
}

func (s *ReceptionService) CreateReception(pvzID, userRole string) (*models.Reception, error) {
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

func (s *ReceptionService) CloseReception(pvzID, userRole string) (*models.Reception, error) {
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

func (s *ReceptionService) DeleteLastProduct(pvzID, userRole string) (*models.Reception, error) {
	if userRole != "employee" {
		return nil, errors.New("access denied")
	}

	active, err := s.repo.GetActiveByPVZ(pvzID)
	if err != nil {
		return nil, err
	}
	if active == nil {
		return nil, errors.New("no active reception")
	}

	return s.repo.DeleteLastProduct(active.ID)
}
