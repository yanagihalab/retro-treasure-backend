package model

type InventoryEntry struct {
	ItemID   int64  `json:"item_id"`
	Name     string `json:"name"`
	Rarity   int    `json:"rarity"`
	Quantity int    `json:"quantity"`
}
