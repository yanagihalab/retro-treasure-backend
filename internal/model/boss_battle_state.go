package model

import "time"

type BossBattleState struct {
	UserID       int64     `json:"user_id"`
	InBattle     bool      `json:"in_battle"`
	BossID       int64     `json:"boss_id"`
	PlayerCardID int64     `json:"player_card_id"`
	PlayerHP     int       `json:"player_hp"`
	BossHP       int       `json:"boss_hp"`
	Turn         int       `json:"turn"`
	LastAction   string    `json:"last_action"`
	UpdatedAt    time.Time `json:"updated_at"`
}
