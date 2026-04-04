package service

import "retro-treasure-backend/internal/repository"

type LoginBonusService struct {
	repo *repository.MemoryRepository
}

func NewLoginBonusService(repo *repository.MemoryRepository) *LoginBonusService {
	return &LoginBonusService{repo: repo}
}

type LoginBonusResponse struct {
	Message     string `json:"message"`
	RewardType  string `json:"reward_type"`
	RewardValue int    `json:"reward_value"`
}

func (s *LoginBonusService) Claim(userID int64) (LoginBonusResponse, error) {
	coins, err := s.repo.ClaimLoginBonus(userID)
	if err != nil {
		return LoginBonusResponse{}, err
	}
	return LoginBonusResponse{
		Message:     "claimed",
		RewardType:  "coin",
		RewardValue: coins,
	}, nil
}
