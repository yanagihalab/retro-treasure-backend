package service

import (
	"retro-treasure-backend/internal/repository"
)

type PlayerService struct {
	repo *repository.MemoryRepository
}

func NewPlayerService(repo *repository.MemoryRepository) *PlayerService {
	return &PlayerService{repo: repo}
}

type PlayerMeResponse struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Level    int    `json:"level"`
	Exp      int    `json:"exp"`
	Coins    int    `json:"coins"`
	Decks    int    `json:"decks"`
	Owned    int    `json:"owned"`
	Tutorial bool   `json:"tutorial"`
}

func (s *PlayerService) GetMe(userID int64) (PlayerMeResponse, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return PlayerMeResponse{}, err
	}

	status, err := s.repo.GetPlayerStatus(userID)
	if err != nil {
		return PlayerMeResponse{}, err
	}

	deck, err := s.repo.ListDeckCards(userID)
	if err != nil {
		return PlayerMeResponse{}, err
	}

	owned, err := s.repo.ListOwnedCards(userID)
	if err != nil {
		return PlayerMeResponse{}, err
	}

	result := PlayerMeResponse{
		UserID:   user.ID,
		Username: user.Username,
		Level:    status.Level,
		Exp:      status.Exp,
		Coins:    status.Coins,
		Decks:    len(deck),
		Owned:    len(owned),
		Tutorial: status.Tutorial,
	}

	if len(owned) >= 4 {
		status.Tutorial = false
		_ = s.repo.SavePlayerStatus(status)
	}

	return result, nil
}
