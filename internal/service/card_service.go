package service

import (
	crand "crypto/rand"
	"math/big"

	"retro-treasure-backend/internal/model"
	"retro-treasure-backend/internal/repository"
)

type CardService struct{ repo *repository.MemoryRepository }

func NewCardService(repo *repository.MemoryRepository) *CardService { return &CardService{repo: repo} }

func (s *CardService) GetMe(userID int64) (model.CardMeResponse, error) {
	card, _, err := s.repo.GetEquippedCard(userID)
	if err != nil {
		return model.CardMeResponse{}, err
	}
	return model.CardMeResponse{EquippedCard: card}, nil
}

func (s *CardService) GetDeck(userID int64) (model.DeckResponse, error) {
	cards, err := s.repo.ListDeckCards(userID)
	if err != nil {
		return model.DeckResponse{}, err
	}
	return model.DeckResponse{Cards: cards}, nil
}

func (s *CardService) GetCollection(userID int64) (model.CardCollectionResponse, error) {
	cards, err := s.repo.ListOwnedCards(userID)
	if err != nil {
		return model.CardCollectionResponse{}, err
	}
	return model.CardCollectionResponse{Cards: cards}, nil
}

func (s *CardService) Upgrade(userID, cardID int64) (model.UpgradeCardResponse, error) {
	card, uc, cost, err := s.repo.UpgradeCard(userID, cardID)
	if err != nil {
		return model.UpgradeCardResponse{}, err
	}
	status, _ := s.repo.GetPlayerStatus(userID)
	return model.UpgradeCardResponse{Card: card, User: uc, Cost: cost, PlayerCoins: status.Coins}, nil
}

func (s *CardService) UpdateDeck(userID int64, cardIDs []int64) (model.DeckResponse, error) {
	if err := s.repo.UpdateDeck(userID, cardIDs); err != nil {
		return model.DeckResponse{}, err
	}
	return s.GetDeck(userID)
}

func (s *CardService) Gacha(userID int64) (model.GachaResponse, error) {
	const cost = 200
	coins, err := s.repo.SpendCoins(userID, cost)
	if err != nil {
		return model.GachaResponse{}, err
	}
	pool := s.repo.ListGachaPool()
	picked := weightedPick(pool)
	card, duplicate, err := s.repo.AddCardToUser(userID, picked.ID)
	if err != nil {
		return model.GachaResponse{}, err
	}
	msg := ""
	if duplicate {
		coins, _ = s.repo.AddCoins(userID, 50)
		msg = "重複カードだったため強化ボーナス + 返還COIN 50 を獲得しました。"
	}
	return model.GachaResponse{SpentCoins: cost, PlayerCoins: coins, Card: card, Duplicate: duplicate, BonusMessage: msg}, nil
}

func weightedPick(pool []model.CharacterCard) model.CharacterCard {
	if len(pool) == 0 {
		return model.CharacterCard{}
	}
	weights := make([]int, len(pool))
	total := 0
	for i, c := range pool {
		w := 100
		switch c.Rarity {
		case 1:
			w = 62
		case 2:
			w = 25
		case 3:
			w = 10
		case 4:
			w = 3
		}
		weights[i] = w
		total += w
	}
	n, _ := crand.Int(crand.Reader, big.NewInt(int64(total)))
	r := int(n.Int64())
	acc := 0
	for i, w := range weights {
		acc += w
		if r < acc {
			return pool[i]
		}
	}
	return pool[0]
}
