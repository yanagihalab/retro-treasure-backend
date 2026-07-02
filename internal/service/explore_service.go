package service

import (
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"retro-treasure-backend/internal/model"
	"retro-treasure-backend/internal/repository"
)

type ExploreService struct {
	repo *repository.MemoryRepository
}

func NewExploreService(repo *repository.MemoryRepository) *ExploreService {
	return &ExploreService{repo: repo}
}

func (s *ExploreService) Explore(userID int64, req model.ExploreRequest) (model.ExploreResult, error) {
	status, err := s.repo.GetPlayerStatus(userID)
	if err != nil {
		return model.ExploreResult{}, err
	}

	area, err := s.repo.GetArea(req.AreaID)
	if err != nil {
		return model.ExploreResult{}, err
	}
	if status.Level < area.RequiredLevel {
		return model.ExploreResult{}, repository.ErrAreaLocked
	}

	drops, err := s.repo.GetDrops(area.ID)
	if err != nil {
		return model.ExploreResult{}, err
	}

	selected, err := chooseDrop(drops)
	if err != nil {
		return model.ExploreResult{}, err
	}

	status.Exp += selected.ExpReward
	status.Coins += selected.CoinReward
	status.TotalExplorations++

	var newItem *model.Item
	registered := false
	var resultItemID *int64
	if selected.ItemID != nil {
		item, err := s.repo.GetItem(*selected.ItemID)
		if err != nil {
			return model.ExploreResult{}, err
		}
		if err := s.repo.AddItemToInventory(userID, item.ID, 1); err != nil {
			return model.ExploreResult{}, err
		}

		isNew, err := s.repo.RegisterEncyclopedia(userID, item.ID)
		if err != nil {
			return model.ExploreResult{}, err
		}

		registered = isNew
		newItem = &item
		resultItemID = &item.ID
	}

	beforeLevel := status.Level
	for status.Exp >= requiredExpForLevel(status.Level+1) {
		status.Level++
	}
	levelUp := status.Level > beforeLevel
	if err := s.repo.SavePlayerStatus(status); err != nil {
		return model.ExploreResult{}, err
	}

	log := model.ExplorationLog{
		UserID:          userID,
		AreaID:          area.ID,
		ConsumedStamina: 0,
		GainedExp:       selected.ExpReward,
		GainedCoins:     selected.CoinReward,
		ResultType:      selected.ResultType,
		ResultItemID:    resultItemID,
		CreatedAt:       time.Now(),
	}
	if err := s.repo.AddExplorationLog(log); err != nil {
		return model.ExploreResult{}, err
	}

	return model.ExploreResult{
		ResultType:             selected.ResultType,
		Message:                selected.Message,
		GainedExp:              selected.ExpReward,
		GainedCoins:            selected.CoinReward,
		LevelUp:                levelUp,
		NewItem:                newItem,
		EncyclopediaRegistered: registered,
		PlayerStatus:           status,
	}, nil
}

func chooseDrop(drops []repository.WeightedDrop) (repository.WeightedDrop, error) {
	if len(drops) == 0 {
		return repository.WeightedDrop{}, errors.New("no drops configured")
	}

	total := 0
	for _, d := range drops {
		total += d.Weight
	}
	if total <= 0 {
		return repository.WeightedDrop{}, errors.New("invalid drop weights")
	}

	n, err := rand.Int(rand.Reader, big.NewInt(int64(total)))
	if err != nil {
		return repository.WeightedDrop{}, err
	}

	pick := int(n.Int64()) + 1
	current := 0
	for _, d := range drops {
		current += d.Weight
		if pick <= current {
			return d, nil
		}
	}
	return drops[len(drops)-1], nil
}

func requiredExpForLevel(level int) int {
	if level <= 1 {
		return 0
	}

	switch level {
	case 2:
		return 10
	case 3:
		return 30
	case 4:
		return 60
	case 5:
		return 100
	case 6:
		return 150
	default:
		return 150 + (level-6)*60
	}
}
