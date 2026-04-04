package model

type CharacterCard struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Rarity        int    `json:"rarity"`
	Element       string `json:"element"`
	MaxHP         int    `json:"max_hp"`
	Attack        int    `json:"attack"`
	Defense       int    `json:"defense"`
	IsStarter     bool   `json:"is_starter"`
	PortraitLabel string `json:"portrait_label"`
	FrameStyle    string `json:"frame_style"`
}
