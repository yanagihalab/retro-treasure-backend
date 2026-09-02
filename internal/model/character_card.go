package model

type CharacterCard struct {
	ID          int64          `json:"id"`
	Name        string         `json:"name"`
	SubName     string         `json:"sub_name"`
	Description string         `json:"description"`
	Rarity      int            `json:"rarity"`
	Element     string         `json:"element"`
	MaxHP       int            `json:"max_hp"`
	Defense     int            `json:"defense"`
	Skills      []DefenseSkill `json:"skills"`
}

type DefenseSkill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	TriggerRate int    `json:"trigger_rate"`
	EffectType  string `json:"effect_type"`
	SubElement  string `json:"sub_element"`
	Heal        int    `json:"heal"`
	Shield      int    `json:"shield"`
	IsEvade     bool   `json:"evade"`
	IsFirst     bool   `json:"first"`
	Target      string `json:"target"`
	IsDamaged   bool   `json:"damaged"`
	IsDead      bool   `json:"dead"`
}
