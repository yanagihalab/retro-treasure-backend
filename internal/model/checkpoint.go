package model

import "time"

type Checkpoint struct {
	ID                string  `json:"id"`
	QRText            string  `json:"qr_text"`
	Name              string  `json:"name"`
	Area              string  `json:"area"`
	Description       string  `json:"description"`
	MapX              int     `json:"map_x"`
	MapY              int     `json:"map_y"`
	Lat               float64 `json:"lat"`
	Lng               float64 `json:"lng"`
	FirstRewardCoin   int     `json:"first_reward_coin"`
	FirstRewardExp    int     `json:"first_reward_exp"`
	DailyRewardCoin   int     `json:"daily_reward_coin"`
	DailyRewardExp    int     `json:"daily_reward_exp"`
	EventRewardName   string  `json:"event_reward_name"`
	EventRewardType   string  `json:"event_reward_type"`
	EventRewardValue  int     `json:"event_reward_value"`
	BossTicketReward  int     `json:"boss_ticket_reward"`
	GachaTicketReward int     `json:"gacha_ticket_reward"`
	IsEventActive     bool    `json:"is_event_active"`
	IsActive          bool    `json:"is_active"`
}

type UserCheckpointRecord struct {
	UserID         int64      `json:"user_id"`
	CheckpointID   string     `json:"checkpoint_id"`
	FirstClaimedAt *time.Time `json:"first_claimed_at,omitempty"`
	LastClaimedAt  *time.Time `json:"last_claimed_at,omitempty"`
	ClaimCount     int        `json:"claim_count"`
	EventClaimed   bool       `json:"event_claimed"`
}

type CheckpointRewardLine struct {
	Type  string `json:"type"`
	Label string `json:"label"`
	Value int    `json:"value"`
}

type CheckpointClaimRequest struct {
	QRText string `json:"qr_text"`
}

type CheckpointClaimResponse struct {
	OK               bool                   `json:"ok"`
	Checkpoint       Checkpoint             `json:"checkpoint"`
	Rewards          []CheckpointRewardLine `json:"rewards"`
	FirstTime        bool                   `json:"first_time"`
	DailyClaim       bool                   `json:"daily_claim"`
	EventClaim       bool                   `json:"event_claim"`
	ClaimCount       int                    `json:"claim_count"`
	PlayerCoins      int                    `json:"player_coins"`
	PlayerExp        int                    `json:"player_exp"`
	GachaTickets     int                    `json:"gacha_tickets"`
	BossChallengeKey int                    `json:"boss_challenge_tickets"`
	Summary          string                 `json:"summary"`
}

type CheckpointHistoryEntry struct {
	Checkpoint    Checkpoint           `json:"checkpoint"`
	Record        UserCheckpointRecord `json:"record"`
	CanClaimDaily bool                 `json:"can_claim_daily"`
}

type CheckpointHistoryResponse struct {
	Entries              []CheckpointHistoryEntry `json:"entries"`
	GachaTickets         int                      `json:"gacha_tickets"`
	BossChallengeTickets int                      `json:"boss_challenge_tickets"`
}

type CheckpointMasterResponse struct {
	Checkpoints []Checkpoint `json:"checkpoints"`
}
