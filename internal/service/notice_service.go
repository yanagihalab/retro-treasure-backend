package service

import (
	"retro-treasure-backend/internal/model"
	"retro-treasure-backend/internal/repository"
)

type NoticeService struct {
	repo *repository.MemoryRepository
}

func NewNoticeService(repo *repository.MemoryRepository) *NoticeService {
	return &NoticeService{repo: repo}
}

func (s *NoticeService) ListNotices() []model.Notice {
	return s.repo.ListNotices()
}
