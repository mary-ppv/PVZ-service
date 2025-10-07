package service

import (
	"PVZ/internal/constants"
	"PVZ/models"
	"PVZ/pkg/metrics"
	"context"
	"errors"
)

type PVZService struct {
	repo PVZRepository
}

func NewPVZService(repo PVZRepository) *PVZService {
	return &PVZService{repo: repo}
}

func (s *PVZService) CreatePVZ(ctx context.Context, name string, city string, userRole string) (*models.PVZ, error) {
	if userRole != "moderator" {
		return nil, errors.New("access denied")
	}

	if city != constants.CityKazan && city != constants.CityMoscow && city != constants.CitySpb {
		return nil, errors.New("invalid city")
	}

	pvz, err := s.repo.CreatePVZ(ctx, name, city)
	if err != nil {
		return nil, errors.New("failed to add product to reception")
	}

	metrics.PVZCreated.Inc()
	return pvz, nil
}

func (s *PVZService) GetPVZList(ctx context.Context, offset, limit int, city string, userRole string) ([]*models.PVZ, error) {
	if userRole != "employee" && userRole != "moderator" {
		return nil, errors.New("access denied")
	}

	list, err := s.repo.GetPVZList(ctx, offset, limit, city)
	if err != nil {
		return nil, errors.New("failed to get PVZ list")
	}

	return list, nil
}
