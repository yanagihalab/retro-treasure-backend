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

type CardArchiveEntry struct {
	Card         CharacterCard `json:"card"`
	Obtained     bool          `json:"obtained"`
	InDeck       bool          `json:"in_deck"`
	DeckSlot     int           `json:"deck_slot,omitempty"`
	UpgradeLevel int           `json:"upgrade_level,omitempty"`
}

type CardArchiveResponse struct {
	Cards          []CardArchiveEntry `json:"cards"`
	Total          int                `json:"total"`
	ObtainedCount  int                `json:"obtained_count"`
	CompletionRate float64            `json:"completion_rate"`
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
	Boss            Boss                  `json:"boss"`
	Bosses          []Boss                `json:"bosses,omitempty"`
	RecommendedDeck []BossRecommendedCard `json:"recommended_deck,omitempty"`
	DropPreview     BossDropPreview       `json:"drop_preview,omitempty"`
}

type BossRecommendedCard struct {
	Card     CharacterCard `json:"card"`
	Reason   string        `json:"reason"`
	Score    int           `json:"score"`
	Owned    bool          `json:"owned"`
	InDeck   bool          `json:"in_deck"`
	DeckSlot int           `json:"deck_slot,omitempty"`
}

type BossDropPreview struct {
	DropRatePercent int             `json:"drop_rate_percent"`
	Candidates      []CharacterCard `json:"candidates"`
}

type BattleLogEntry struct {
	Round             int    `json:"round"`
	ActorType         string `json:"actor_type"`
	ActorName         string `json:"actor_name"`
	TargetName        string `json:"target_name"`
	Action            string `json:"action"`
	SkillName         string `json:"skill_name,omitempty"`
	EffectType        string `json:"effect_type,omitempty"`
	DefenseSkillName  string `json:"defense_skill_name,omitempty"`
	DefenseSkillTier  string `json:"defense_skill_tier,omitempty"`
	DefenseEffectType string `json:"defense_effect_type,omitempty"`
	DamageReduced     int    `json:"damage_reduced,omitempty"`
	HealAmount        int    `json:"heal_amount,omitempty"`
	SupportTargetName string `json:"support_target_name,omitempty"`
	SupportTargetSlot int    `json:"support_target_slot"`
	SupportHealAmount int    `json:"support_heal_amount,omitempty"`
	Revived           bool   `json:"revived,omitempty"`
	Evaded            bool   `json:"evaded,omitempty"`
	Damage            int    `json:"damage"`
	TargetHP          int    `json:"target_hp"`
	TargetMaxHP       int    `json:"target_max_hp"`
	CardSlot          int    `json:"card_slot,omitempty"`
	TargetSlot        int    `json:"target_slot,omitempty"`
	CardHPAfter       int    `json:"card_hp_after,omitempty"`
	BossHP            int    `json:"boss_hp,omitempty"`
	Message           string `json:"message"`
	TargetKO          bool   `json:"target_ko"`
}

type BattleReward struct {
	Exp               int            `json:"exp"`
	Coins             int            `json:"coins"`
	RewardCard        *CharacterCard `json:"reward_card,omitempty"`
	BossDropCard      *CharacterCard `json:"boss_drop_card,omitempty"`
	BossDropRate      int            `json:"boss_drop_rate,omitempty"`
	BossDropOccurred  bool           `json:"boss_drop_occurred,omitempty"`
	BossDropDuplicate bool           `json:"boss_drop_duplicate,omitempty"`
	BossDropMessage   string         `json:"boss_drop_message,omitempty"`
}

type BattleDeckCardState struct {
	Card       CharacterCard `json:"card"`
	DeckSlot   int           `json:"deck_slot"`
	CurrentHP  int           `json:"current_hp"`
	MaxHP      int           `json:"max_hp"`
	KnockedOut bool          `json:"knocked_out"`
}

type AutoBattleResponse struct {
	BattleTitle   string                `json:"battle_title"`
	BattleMode    string                `json:"battle_mode,omitempty"`
	TargetTurns   int                   `json:"target_turns,omitempty"`
	SurvivedTurns int                   `json:"survived_turns,omitempty"`
	Boss          Boss                  `json:"boss"`
	InitialDeck   []BattleDeckCardState `json:"initial_deck"`
	Deck          []BattleDeckCardState `json:"deck"`
	Logs          []BattleLogEntry      `json:"logs"`
	Result        string                `json:"result"`
	Reward        BattleReward          `json:"reward"`
	Summary       string                `json:"summary"`
}
