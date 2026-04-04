package model

type Area struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	RequiredLevel int    `json:"required_level"`
	StaminaCost   int    `json:"stamina_cost"`
	IsActive      bool   `json:"is_active"`
	SortOrder     int    `json:"sort_order"`
}
