package model

type CardMeResponse struct {
	EquippedCard CharacterCard `json:"equipped_card"`
}

type DeckCardView struct {
	Card         CharacterCard `json:"card"`
	DeckSlot     int           `json:"deck_slot"`
	UpgradeLevel int           `json:"upgrade_level"`
}

type DeckResponse struct {
	Cards []DeckCardView `json:"cards"`
}

type CardCollectionEntry struct {
	Card   CharacterCard     `json:"card"`
	User   UserCharacterCard `json:"user"`
	InDeck bool              `json:"in_deck"`
}

type CardCollectionResponse struct {
	Cards []CardCollectionEntry `json:"cards"`
}

type UpgradeCardRequest struct {
	CardID int64 `json:"card_id"`
}

type UpgradeCardResponse struct {
	Card        CharacterCard     `json:"card"`
	User        UserCharacterCard `json:"user"`
	Cost        int               `json:"cost"`
	PlayerCoins int               `json:"player_coins"`
}

type UpdateDeckRequest struct {
	CardIDs []int64 `json:"card_ids"`
}

type GachaResponse struct {
	SpentCoins   int           `json:"spent_coins"`
	PlayerCoins  int           `json:"player_coins"`
	Card         CharacterCard `json:"card"`
	Duplicate    bool          `json:"duplicate"`
	BonusMessage string        `json:"bonus_message,omitempty"`
}

type BossInfoResponse struct {
	Boss Boss `json:"boss"`
}

type BattleLogEntry struct {
	Round       int    `json:"round"`
	ActorType   string `json:"actor_type"`
	ActorName   string `json:"actor_name"`
	TargetName  string `json:"target_name"`
	Action      string `json:"action"`
	Damage      int    `json:"damage"`
	TargetHP    int    `json:"target_hp"`
	TargetMaxHP int    `json:"target_max_hp"`
	CardSlot    int    `json:"card_slot,omitempty"`
	TargetSlot  int    `json:"target_slot,omitempty"`
	CardHPAfter int    `json:"card_hp_after,omitempty"`
	BossHP      int    `json:"boss_hp,omitempty"`
	Message     string `json:"message"`
	TargetKO    bool   `json:"target_ko"`
}

type BattleReward struct {
	Exp        int            `json:"exp"`
	Coins      int            `json:"coins"`
	RewardCard *CharacterCard `json:"reward_card,omitempty"`
}

type BattleDeckCardState struct {
	Card       CharacterCard `json:"card"`
	DeckSlot   int           `json:"deck_slot"`
	CurrentHP  int           `json:"current_hp"`
	MaxHP      int           `json:"max_hp"`
	KnockedOut bool          `json:"knocked_out"`
}

type AutoBattleResponse struct {
	BattleTitle string                `json:"battle_title"`
	Boss        Boss                  `json:"boss"`
	InitialDeck []BattleDeckCardState `json:"initial_deck"`
	Deck        []BattleDeckCardState `json:"deck"`
	Logs        []BattleLogEntry      `json:"logs"`
	Result      string                `json:"result"`
	Reward      BattleReward          `json:"reward"`
	Summary     string                `json:"summary"`
}
