package service

import (
	"PVZ/models"
	"PVZ/pkg/logger"
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
		logger.Log.Printf("Access denied: userRole=%s tried to create PVZ %s", userRole, name)
		return nil, errors.New("access denied")
	}

	if city != models.CityMoscow && city != models.CitySpb && city != models.CityKazan {
		logger.Log.Printf("Invalid city: %s for PVZ %s", city, name)
		return nil, errors.New("invalid city")
	}

	pvz, err := s.repo.CreatePVZ(ctx, name, city)
	if err != nil {
		logger.Log.Printf("Failed to create PVZ %s in city %s: %v", name, city, err)
		return nil, err
	}

	logger.Log.Printf("PVZ created: %s in city %s", name, city)
	metrics.PVZCreated.Inc()
	return pvz, nil
}

func (s *PVZService) GetPVZList(ctx context.Context, offset, limit int, city string, userRole string) ([]*models.PVZ, error) {
	if userRole != "employee" && userRole != "moderator" {
		logger.Log.Printf("Access denied: userRole=%s tried to list PVZ", userRole)
		return nil, errors.New("access denied")
	}

	list, err := s.repo.GetPVZList(ctx, offset, limit, city)
	if err != nil {
		logger.Log.Printf("Failed to get PVZ list: %v", err)
		return nil, err
	}

	logger.Log.Printf("PVZ list fetched by userRole=%s, count=%d", userRole, len(list))
	return list, nil
}
