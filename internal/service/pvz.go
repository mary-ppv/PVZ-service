package service

import (
	"PVZ/internal/models"
	"PVZ/pkg/metrics"
	"errors"
)

type PVZService struct {
	repo PVZRepository
}

func NewPVZService(repo PVZRepository) *PVZService {
	return &PVZService{repo: repo}
}

func (s *PVZService) CreatePVZ(name string, city models.City, userRole string) (*models.PVZ, error) {
	if userRole != "moderator" {
		return nil, errors.New("access denied")
	}

	if city != models.CityMoscow && city != models.CitySpb && city != models.CityKazan {
		return nil, errors.New("invalid city")
	}

	metrics.PVZCreated.Inc()
	return s.repo.CreatePVZ(name, city)
}

func (s *PVZService) GetPVZList(offset, limit int, city *models.City, userRole string) ([]*models.PVZ, error) {
	if userRole != "employee" && userRole != "moderator" {
		return nil, errors.New("access denied")
	}

	return s.repo.GetPVZList(offset, limit, city)
}
