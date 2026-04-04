package model

import "time"

type UserCharacterCard struct {
	UserID       int64     `json:"user_id"`
	CardID       int64     `json:"card_id"`
	IsEquipped   bool      `json:"is_equipped"`
	DeckSlot     int       `json:"deck_slot"`
	Level        int       `json:"level"`
	BonusHP      int       `json:"bonus_hp"`
	BonusAttack  int       `json:"bonus_attack"`
	BonusDefense int       `json:"bonus_defense"`
	AcquiredAt   time.Time `json:"acquired_at"`
}
