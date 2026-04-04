package model

type Boss struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Element       string `json:"element"`
	MaxHP         int    `json:"max_hp"`
	Attack        int    `json:"attack"`
	Defense       int    `json:"defense"`
	RewardExp     int    `json:"reward_exp"`
	RewardCoins   int    `json:"reward_coins"`
	PortraitLabel string `json:"portrait_label"`
	FrameStyle    string `json:"frame_style"`
}
