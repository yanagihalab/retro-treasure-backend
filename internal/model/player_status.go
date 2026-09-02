package model

import "time"

type PlayerStatus struct {
	UserID    int64     `json:"user_id"`
	Level     int       `json:"level"`
	Exp       int       `json:"exp"`
	Coins     int       `json:"coins"`
	Tutorial  bool      `json:"tutorial"`
	UpdatedAt time.Time `json:"updated_at"`
}
