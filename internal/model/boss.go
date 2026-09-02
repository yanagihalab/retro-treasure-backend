package model

type Boss struct {
	ID          int64              `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	RewardExp   int                `json:"reward_exp"`
	RewardCoins int                `json:"reward_coins"`
	AttackMoves [][]BossAttackMove `json:"attack_moves"`
	DropRate    int                `json:"drop_rate"`
	Candidates  []int64            `json:"candidates"`
}

type BossAttackMove struct {
	Name       string `json:"name"`
	Element    string `json:"element"`
	SubElement string `json:"sub_element"`
	Power      int    `json:"power"`
	All        bool   `json:"all"`
}
