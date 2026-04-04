package model

import "time"

type PlayerStatus struct {
	UserID                 int64     `json:"user_id"`
	Level                  int       `json:"level"`
	Exp                    int       `json:"exp"`
	Stamina                int       `json:"stamina"`
	MaxStamina             int       `json:"max_stamina"`
	Coins                  int       `json:"coins"`
	Gems                   int       `json:"gems"`
	TotalExplorations      int       `json:"total_explorations"`
	LastStaminaRecoveredAt time.Time `json:"last_stamina_recovered_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}
