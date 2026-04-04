package service

import (
	"retro-treasure-backend/internal/model"
	"retro-treasure-backend/internal/repository"
)

type ItemService struct {
	repo *repository.MemoryRepository
}

func NewItemService(repo *repository.MemoryRepository) *ItemService {
	return &ItemService{repo: repo}
}

func (s *ItemService) ListInventory(userID int64) ([]model.InventoryEntry, error) {
	return s.repo.ListInventory(userID)
}

type EncyclopediaResponse struct {
	CompletionRate float64                   `json:"completion_rate"`
	Entries        []model.EncyclopediaEntry `json:"entries"`
}

func (s *ItemService) GetEncyclopedia(userID int64) (EncyclopediaResponse, error) {
	entries, err := s.repo.ListEncyclopedia(userID)
	if err != nil {
		return EncyclopediaResponse{}, err
	}
	if len(entries) == 0 {
		return EncyclopediaResponse{CompletionRate: 0, Entries: entries}, nil
	}
	obtained := 0
	for _, e := range entries {
		if e.Obtained {
			obtained++
		}
	}
	rate := float64(obtained) * 100 / float64(len(entries))
	return EncyclopediaResponse{CompletionRate: rate, Entries: entries}, nil
}
