package service

import (
	"retro-treasure-backend/internal/model"
	"retro-treasure-backend/internal/repository"
)

type AreaService struct {
	repo *repository.MemoryRepository
}

func NewAreaService(repo *repository.MemoryRepository) *AreaService {
	return &AreaService{repo: repo}
}

func (s *AreaService) ListAreas() []model.Area {
	return s.repo.ListAreas()
}
