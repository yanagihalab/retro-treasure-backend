package service

import (
	"strings"
	"time"

	"retro-treasure-backend/internal/model"
	"retro-treasure-backend/internal/repository"
)

type CheckpointService struct {
	repo *repository.MemoryRepository
}

func NewCheckpointService(repo *repository.MemoryRepository) *CheckpointService {
	return &CheckpointService{repo: repo}
}

func (s *CheckpointService) GetMaster() model.CheckpointMasterResponse {
	return model.CheckpointMasterResponse{Checkpoints: s.repo.ListCheckpoints()}
}

func (s *CheckpointService) GetHistory(userID int64) model.CheckpointHistoryResponse {
	checkpoints := s.repo.ListCheckpoints()
	entries := make([]model.CheckpointHistoryEntry, 0, len(checkpoints))
	now := time.Now()

	for _, cp := range checkpoints {
		rec := s.repo.GetCheckpointRecord(userID, cp.ID)
		canDaily := rec.LastClaimedAt == nil || !sameDateLocal(*rec.LastClaimedAt, now)
		entries = append(entries, model.CheckpointHistoryEntry{
			Checkpoint:    cp,
			Record:        rec,
			CanClaimDaily: canDaily,
		})
	}

	return model.CheckpointHistoryResponse{
		Entries:              entries,
		GachaTickets:         s.repo.GetGachaTickets(userID),
		BossChallengeTickets: s.repo.GetBossChallengeTickets(userID),
	}
}

func (s *CheckpointService) Claim(userID int64, qrText string) (model.CheckpointClaimResponse, error) {
	cp, err := s.repo.GetCheckpointByQR(strings.TrimSpace(qrText))
	if err != nil {
		return model.CheckpointClaimResponse{}, err
	}
	now := time.Now()
	rec := s.repo.GetCheckpointRecord(userID, cp.ID)
	status, err := s.repo.GetPlayerStatus(userID)
	if err != nil {
		return model.CheckpointClaimResponse{}, err
	}

	firstTime := rec.ClaimCount == 0
	dailyClaim := rec.LastClaimedAt == nil || !sameDateLocal(*rec.LastClaimedAt, now)
	eventClaim := false
	rewards := make([]model.CheckpointRewardLine, 0, 8)

	if firstTime {
		if cp.FirstRewardCoin > 0 {
			status.Coins += cp.FirstRewardCoin
			rewards = append(rewards, model.CheckpointRewardLine{
				Type:  "coin",
				Label: "初回通過報酬",
				Value: cp.FirstRewardCoin,
			})
		}
		if cp.FirstRewardExp > 0 {
			_, _ = s.repo.AddExp(userID, cp.FirstRewardExp)
			rewards = append(rewards, model.CheckpointRewardLine{
				Type:  "exp",
				Label: "初回通過報酬",
				Value: cp.FirstRewardExp,
			})
		}
		if cp.GachaTicketReward > 0 {
			s.repo.AddGachaTickets(userID, cp.GachaTicketReward)
			rewards = append(rewards, model.CheckpointRewardLine{
				Type:  "gacha_ticket",
				Label: "ガチャチケット",
				Value: cp.GachaTicketReward,
			})
		}
		if cp.BossTicketReward > 0 {
			s.repo.AddBossChallengeTickets(userID, cp.BossTicketReward)
			rewards = append(rewards, model.CheckpointRewardLine{
				Type:  "boss_ticket",
				Label: "ボス挑戦権",
				Value: cp.BossTicketReward,
			})
		}
	}
	if dailyClaim {
		if cp.DailyRewardCoin > 0 {
			status.Coins += cp.DailyRewardCoin
			rewards = append(rewards, model.CheckpointRewardLine{
				Type:  "coin",
				Label: "日次報酬",
				Value: cp.DailyRewardCoin,
			})
		}
		if cp.DailyRewardExp > 0 {
			_, _ = s.repo.AddExp(userID, cp.DailyRewardExp)
			rewards = append(rewards, model.CheckpointRewardLine{
				Type:  "exp",
				Label: "日次報酬",
				Value: cp.DailyRewardExp,
			})
		}
	}
	if cp.IsEventActive && !rec.EventClaimed && cp.EventRewardType != "" {
		eventClaim = true
		rec.EventClaimed = true
		label := cp.EventRewardName
		if label == "" {
			label = "イベント限定報酬"
		}
		switch cp.EventRewardType {
		case "coin":
			status.Coins += cp.EventRewardValue
		case "gacha_ticket":
			s.repo.AddGachaTickets(userID, cp.EventRewardValue)
		case "boss_ticket":
			s.repo.AddBossChallengeTickets(userID, cp.EventRewardValue)
		}
		rewards = append(rewards, model.CheckpointRewardLine{
			Type:  cp.EventRewardType,
			Label: label,
			Value: cp.EventRewardValue,
		})
	}
	if len(rewards) == 0 {
		return s.claimResponseWithoutRewards(userID, cp, rec, rewards, status), nil
	}

	rec.ClaimCount++
	rec.UserID = userID
	rec.CheckpointID = cp.ID
	if rec.FirstClaimedAt == nil {
		t := now
		rec.FirstClaimedAt = &t
	}
	t := now
	rec.LastClaimedAt = &t
	_ = s.repo.SaveCheckpointRecord(rec)
	_ = s.repo.SavePlayerStatus(status)
	st2, _ := s.repo.GetPlayerStatus(userID)
	summary := "チェックポイント報酬を獲得しました。"
	if firstTime {
		summary = "初回通過報酬を獲得しました。"
	} else if dailyClaim {
		summary = "日次報酬を獲得しました。"
	}

	return s.claimResponseWithRewards(
		userID,
		cp,
		rec,
		rewards,
		st2,
		firstTime,
		dailyClaim,
		eventClaim,
		summary,
	), nil
}

func (s *CheckpointService) claimResponseWithoutRewards(
	userID int64,
	cp model.Checkpoint,
	rec model.UserCheckpointRecord,
	rewards []model.CheckpointRewardLine,
	status model.PlayerStatus,
) model.CheckpointClaimResponse {
	return model.CheckpointClaimResponse{
		OK:               true,
		Checkpoint:       cp,
		Rewards:          rewards,
		FirstTime:        false,
		DailyClaim:       false,
		EventClaim:       false,
		ClaimCount:       rec.ClaimCount,
		PlayerCoins:      status.Coins,
		PlayerExp:        status.Exp,
		GachaTickets:     s.repo.GetGachaTickets(userID),
		BossChallengeKey: s.repo.GetBossChallengeTickets(userID),
		Summary:          "本日の報酬は受取済みです。",
	}
}

func (s *CheckpointService) claimResponseWithRewards(
	userID int64,
	cp model.Checkpoint,
	rec model.UserCheckpointRecord,
	rewards []model.CheckpointRewardLine,
	status model.PlayerStatus,
	firstTime bool,
	dailyClaim bool,
	eventClaim bool,
	summary string,
) model.CheckpointClaimResponse {
	return model.CheckpointClaimResponse{
		OK:               true,
		Checkpoint:       cp,
		Rewards:          rewards,
		FirstTime:        firstTime,
		DailyClaim:       dailyClaim,
		EventClaim:       eventClaim,
		ClaimCount:       rec.ClaimCount,
		PlayerCoins:      status.Coins,
		PlayerExp:        status.Exp,
		GachaTickets:     s.repo.GetGachaTickets(userID),
		BossChallengeKey: s.repo.GetBossChallengeTickets(userID),
		Summary:          summary,
	}
}

func sameDateLocal(a, b time.Time) bool {
	y1, m1, d1 := a.Date()
	y2, m2, d2 := b.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
