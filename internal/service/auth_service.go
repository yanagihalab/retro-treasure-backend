package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

	"retro-treasure-backend/internal/model"
	"retro-treasure-backend/internal/repository"
)

type AuthService struct {
	repo *repository.MemoryRepository
}

func NewAuthService(repo *repository.MemoryRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) Register(req model.RegisterRequest) (model.AuthResponse, error) {
	if strings.TrimSpace(req.Username) == "" || len(req.Password) < 8 {
		return model.AuthResponse{}, errors.New("username is required and password must be at least 8 chars")
	}

	hash := hashPassword(req.Password)
	user, _, err := s.repo.CreateUser(strings.TrimSpace(req.Username), hash)
	if err != nil {
		return model.AuthResponse{}, err
	}
	token, _, err := s.repo.Login(user.Username, hash)
	if err != nil {
		return model.AuthResponse{}, err
	}
	return model.AuthResponse{Token: token, UserID: user.ID}, nil
}

func (s *AuthService) Login(req model.LoginRequest) (model.AuthResponse, error) {
	hash := hashPassword(req.Password)
	token, user, err := s.repo.Login(strings.TrimSpace(req.Username), hash)
	if err != nil {
		return model.AuthResponse{}, err
	}
	return model.AuthResponse{Token: token, UserID: user.ID}, nil
}

func hashPassword(input string) string {
	sum := sha256.Sum256([]byte(input))
	return hex.EncodeToString(sum[:])
}
