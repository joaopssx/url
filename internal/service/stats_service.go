package service

import (
	"urlshortener/internal/model"
	"urlshortener/internal/repository"
)

type StatsService interface {
	GetStats(code string) (*model.StatsResult, error)
}

type statsService struct {
	urlRepo repository.URLRepository
}

func NewStatsService(urlRepo repository.URLRepository) StatsService {
	return &statsService{urlRepo: urlRepo}
}

func (s *statsService) GetStats(code string) (*model.StatsResult, error) {
	u, err := s.urlRepo.FindByCode(code)
	if err != nil {
		return nil, err
	}
	if u == nil || u.DeletedAt != nil {
		return nil, ErrNotFound
	}

	daily, weekly, monthly, err := s.urlRepo.GetAccessStats(u.ID)
	if err != nil {
		return nil, err
	}

	return &model.StatsResult{
		ShortCode:     u.ShortCode,
		OriginalURL:   u.OriginalURL,
		TotalAccesses: u.AccessCount,
		CreatedAt:     u.CreatedAt,
		ExpiresAt:     u.ExpiresAt,
		Daily:         daily,
		Weekly:        weekly,
		Monthly:       monthly,
	}, nil
}
