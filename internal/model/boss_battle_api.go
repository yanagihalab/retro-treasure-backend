package model

type DeckCardView struct {
	Card     CharacterCard `json:"card"`
	DeckSlot int           `json:"deck_slot"`
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
	Card     CharacterCard `json:"card"`
	Obtained bool          `json:"obtained"`
	InDeck   bool          `json:"in_deck"`
	DeckSlot int           `json:"deck_slot,omitempty"`
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

// ボス一覧ページ
type BossInfoResponse struct {
	Boss         Boss             `json:"boss"`
	StrategyHint BossStrategyHint `json:"strategy_hint"`
	Drops        []BossDrop       `json:"drops"`
}

type BossStrategyHint struct {
	EffectiveElements []string `json:"effective_elements,omitempty"`
	DangerousMoves    []string `json:"dangerous_moves,omitempty"`
	RecommendedCards  []string `json:"recommended_cards,omitempty"`
}

type BossDrop struct {
	DropRate   int            `json:"drop_rate"`
	Candidates []BossDropItem `json:"candidates"`
}

type BossDropItem struct {
	Owned bool          `json:"owned"`
	Card  CharacterCard `json:"card"`
}

type BattleLogEntry struct {
	Round           int    `json:"round"`
	SkillIndex      int    `json:"skill_index"`
	ActorType       string `json:"actor_type"`  // ターン、攻撃、スキルなど
	ActorSlot       int    `json:"actor_slot"`  // スキル使用者のスロット（スキル使用時のみ）
	TargetSlot      int    `json:"target_slot"` // ターゲットのスロット
	CardHPAfter     int    `json:"card_hp_after"`
	CardShieldAfter int    `json:"card_shield_after"`
	CardEvadeAfter  bool   `json:"card_evade_after"`
	Damage          int    `json:"damage"`
	Index           []int  `json:"index"`
}

// バトルページ
type AutoBattleResponse struct {
	Boss        Boss                  `json:"boss"`
	InitialDeck map[int]CharacterCard `json:"initial_deck"`
	Logs        []BattleLogEntry      `json:"logs"`
	IsWin       bool                  `json:"win"`
}

type AutoResultResponse struct {
	Exp        int            `json:"exp"`
	Coins      int            `json:"coins"`
	DropDice   int            `json:"drop_dice"`
	Dropped    bool           `json:"dropped"`
	Duplicate  bool           `json:"duplicate"`
	RewardCard *CharacterCard `json:"reward_card"`
}
