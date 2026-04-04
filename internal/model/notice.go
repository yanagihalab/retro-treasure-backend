package model

import "time"

type Notice struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	IsPinned    bool      `json:"is_pinned"`
	PublishedAt time.Time `json:"published_at"`
	IsActive    bool      `json:"is_active"`
}
