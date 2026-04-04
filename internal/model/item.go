package model

type Item struct {
	ID                   int64  `json:"id"`
	Name                 string `json:"name"`
	Description          string `json:"description"`
	Rarity               int    `json:"rarity"`
	ItemType             string `json:"item_type"`
	SellPrice            int    `json:"sell_price"`
	IsEncyclopediaTarget bool   `json:"is_encyclopedia_target"`
	IsEventLimited       bool   `json:"is_event_limited"`
	IconPath             string `json:"icon_path"`
}
