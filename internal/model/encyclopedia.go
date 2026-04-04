package model

import "time"

type EncyclopediaEntry struct {
	ItemID          int64      `json:"item_id"`
	Name            string     `json:"name"`
	Rarity          int        `json:"rarity"`
	Obtained        bool       `json:"obtained"`
	FirstObtainedAt *time.Time `json:"first_obtained_at,omitempty"`
}
