package model

type Boss struct {
	ID            int64            `json:"id"`
	Name          string           `json:"name"`
	Description   string           `json:"description"`
	Element       string           `json:"element"`
	MaxHP         int              `json:"max_hp"`
	CurrentHP     *int             `json:"current_hp,omitempty"`
	Attack        int              `json:"attack"`
	Defense       int              `json:"defense"`
	RewardExp     int              `json:"reward_exp"`
	RewardCoins   int              `json:"reward_coins"`
	PortraitLabel string           `json:"portrait_label"`
	FrameStyle    string           `json:"frame_style"`
	AttackMoves   []BossAttackMove `json:"attack_moves,omitempty"`
	StrategyHint  BossStrategyHint `json:"strategy_hint,omitempty"`
}

type BossAttackMove struct {
	Name       string `json:"name"`
	EffectType string `json:"effect_type"`
	PowerBonus int    `json:"power_bonus"`
}

type BossStrategyHint struct {
	EffectiveElements []string `json:"effective_elements,omitempty"`
	DangerousMoves    []string `json:"dangerous_moves,omitempty"`
	RecommendedCards  []string `json:"recommended_cards,omitempty"`
}
