package model

import "time"

type ExplorationLog struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	AreaID          int64     `json:"area_id"`
	ConsumedStamina int       `json:"consumed_stamina"`
	GainedExp       int       `json:"gained_exp"`
	GainedCoins     int       `json:"gained_coins"`
	ResultType      string    `json:"result_type"`
	ResultItemID    *int64    `json:"result_item_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type ExploreResult struct {
	ResultType             string       `json:"result_type"`
	Message                string       `json:"message"`
	GainedExp              int          `json:"gained_exp"`
	GainedCoins            int          `json:"gained_coins"`
	LevelUp                bool         `json:"level_up"`
	NewItem                *Item        `json:"new_item,omitempty"`
	EncyclopediaRegistered bool         `json:"encyclopedia_registered"`
	PlayerStatus           PlayerStatus `json:"player_status"`
}
