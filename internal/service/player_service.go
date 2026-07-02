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
	UserID            int64  `json:"user_id"`
	Username          string `json:"username"`
	Level             int    `json:"level"`
	Exp               int    `json:"exp"`
	Stamina           int    `json:"stamina"`
	MaxStamina        int    `json:"max_stamina"`
	StaminaDisplay    string `json:"stamina_display"`
	StaminaInfinite   bool   `json:"stamina_infinite"`
	Coins             int    `json:"coins"`
	Gems              int    `json:"gems"`
	TotalExplorations int    `json:"total_explorations"`
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

	return PlayerMeResponse{
		UserID:            user.ID,
		Username:          user.Username,
		Level:             status.Level,
		Exp:               status.Exp,
		Stamina:           status.Stamina,
		MaxStamina:        status.MaxStamina,
		StaminaDisplay:    "∞",
		StaminaInfinite:   true,
		Coins:             status.Coins,
		Gems:              status.Gems,
		TotalExplorations: status.TotalExplorations,
	}, nil
}
