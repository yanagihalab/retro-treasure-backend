package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

	"retro-treasure-backend/internal/model"
	"retro-treasure-backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
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

	hash, err := hashPassword(req.Password)
	if err != nil {
		return model.AuthResponse{}, err
	}
	user, _, err := s.repo.CreateUser(strings.TrimSpace(req.Username), hash)
	if err != nil {
		return model.AuthResponse{}, err
	}
	token, _, err := s.repo.IssueToken(user.ID)
	if err != nil {
		return model.AuthResponse{}, err
	}
	return model.AuthResponse{Token: token, UserID: user.ID}, nil
}

func (s *AuthService) Login(req model.LoginRequest) (model.AuthResponse, error) {
	user, err := s.repo.GetUserByUsername(strings.TrimSpace(req.Username))
	if err != nil {
		return model.AuthResponse{}, err
	}
	if !verifyPassword(req.Password, user.PasswordHash) {
		return model.AuthResponse{}, repository.ErrInvalidCredentials
	}
	if isLegacySHA256Hash(user.PasswordHash) {
		hash, err := hashPassword(req.Password)
		if err == nil {
			_ = s.repo.UpdatePasswordHash(user.ID, hash)
		}
	}
	token, _, err := s.repo.IssueToken(user.ID)
	if err != nil {
		return model.AuthResponse{}, err
	}
	return model.AuthResponse{Token: token, UserID: user.ID}, nil
}

func hashPassword(input string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(input), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func verifyPassword(input, storedHash string) bool {
	if strings.HasPrefix(storedHash, "$2a$") ||
		strings.HasPrefix(storedHash, "$2b$") ||
		strings.HasPrefix(storedHash, "$2y$") {
		return bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(input)) == nil
	}
	return storedHash == legacySHA256Hash(input)
}

func isLegacySHA256Hash(storedHash string) bool {
	return len(storedHash) == 64 && !strings.HasPrefix(storedHash, "$2")
}

func legacySHA256Hash(input string) string {
	sum := sha256.Sum256([]byte(input))
	return hex.EncodeToString(sum[:])
}
