package model

import "time"

type UserCharacterCard struct {
	UserID     int64     `json:"user_id"`
	CardID     int64     `json:"card_id"`
	DeckSlot   int       `json:"deck_slot"`
	AcquiredAt time.Time `json:"acquired_at"`
}
