package service

import (
	crand "crypto/rand"
	"fmt"
	"math"
	"math/big"

	"retro-treasure-backend/internal/model"
	"retro-treasure-backend/internal/repository"
)

type BossService struct{ repo *repository.MemoryRepository }

func NewBossService(repo *repository.MemoryRepository) *BossService { return &BossService{repo: repo} }

type battleCardState struct {
	Card      model.CharacterCard
	DeckSlot  int
	CurrentHP int
}

func (s *BossService) GetBoss() (model.BossInfoResponse, error) {
	boss, err := s.repo.GetBoss(1)
	if err != nil {
		return model.BossInfoResponse{}, err
	}
	return model.BossInfoResponse{Boss: boss}, nil
}

func (s *BossService) AutoBattle(userID int64) (model.AutoBattleResponse, error) {
	boss, err := s.repo.GetBoss(1)
	if err != nil {
		return model.AutoBattleResponse{}, err
	}
	deckView, err := s.repo.ListDeckCards(userID)
	if err != nil {
		return model.AutoBattleResponse{}, err
	}
	deck := make([]battleCardState, 0, len(deckView))
	for _, dc := range deckView {
		deck = append(deck, battleCardState{Card: dc.Card, DeckSlot: dc.DeckSlot, CurrentHP: dc.Card.MaxHP})
	}
	initialDeck := make([]model.BattleDeckCardState, 0, len(deck))
	for _, dc := range deck {
		initialDeck = append(initialDeck, model.BattleDeckCardState{Card: dc.Card, DeckSlot: dc.DeckSlot, CurrentHP: dc.CurrentHP, MaxHP: dc.Card.MaxHP})
	}
	bossHP := boss.MaxHP
	logs := make([]model.BattleLogEntry, 0, 64)
	round := 1
	for bossHP > 0 && anyAlive(deck) {
		for i := range deck {
			if deck[i].CurrentHP <= 0 || bossHP <= 0 {
				continue
			}
			dmg := calcDamage(deck[i].Card, boss)
			bossHP -= dmg
			if bossHP < 0 {
				bossHP = 0
			}
			logs = append(logs, model.BattleLogEntry{Round: round, ActorType: "card", ActorName: deck[i].Card.Name, TargetName: boss.Name, Action: "attack", Damage: dmg, TargetHP: bossHP, TargetMaxHP: boss.MaxHP, CardSlot: deck[i].DeckSlot - 1, BossHP: bossHP, Message: fmt.Sprintf("%s の攻撃！ %s に %d ダメージ！", deck[i].Card.Name, boss.Name, dmg), TargetKO: bossHP == 0})
		}
		if bossHP == 0 {
			break
		}
		targetIdx := randomAliveIndex(deck)
		if targetIdx < 0 {
			break
		}
		bossDmg := calcBossDamage(boss, deck[targetIdx].Card)
		deck[targetIdx].CurrentHP -= bossDmg
		if deck[targetIdx].CurrentHP < 0 {
			deck[targetIdx].CurrentHP = 0
		}
		logs = append(logs, model.BattleLogEntry{Round: round, ActorType: "boss", ActorName: boss.Name, TargetName: deck[targetIdx].Card.Name, Action: "smash", Damage: bossDmg, TargetHP: deck[targetIdx].CurrentHP, TargetMaxHP: deck[targetIdx].Card.MaxHP, TargetSlot: deck[targetIdx].DeckSlot - 1, CardHPAfter: deck[targetIdx].CurrentHP, Message: fmt.Sprintf("%s の石拳！ %s に %d ダメージ！", boss.Name, deck[targetIdx].Card.Name, bossDmg), TargetKO: deck[targetIdx].CurrentHP == 0})
		round++
	}
	respDeck := make([]model.BattleDeckCardState, 0, len(deck))
	for _, dc := range deck {
		respDeck = append(respDeck, model.BattleDeckCardState{Card: dc.Card, DeckSlot: dc.DeckSlot, CurrentHP: dc.CurrentHP, MaxHP: dc.Card.MaxHP, KnockedOut: dc.CurrentHP <= 0})
	}
	resp := model.AutoBattleResponse{BattleTitle: "AUTO BATTLE", Boss: boss, InitialDeck: initialDeck, Deck: respDeck, Logs: logs}
	if bossHP == 0 {
		resp.Result = "win"
		resp.Reward.Exp = boss.RewardExp
		resp.Reward.Coins = boss.RewardCoins
		resp.Summary = fmt.Sprintf("%s を撃破！ EXP %d / COIN %d を獲得！", boss.Name, boss.RewardExp, boss.RewardCoins)
		status, _ := s.repo.GetPlayerStatus(userID)
		status.Exp += boss.RewardExp
		status.Coins += boss.RewardCoins
		for status.Exp >= requiredExpForLevel(status.Level+1) {
			status.Level++
		}
		_ = s.repo.SavePlayerStatus(status)
		if reward, ok := s.pickRewardCard(userID); ok {
			resp.Reward.RewardCard = &reward
		}
	} else {
		resp.Result = "lose"
		resp.Reward.Coins = 20
		resp.Summary = fmt.Sprintf("%s に敗北... しかし COIN %d を持ち帰った。", boss.Name, resp.Reward.Coins)
		status, _ := s.repo.GetPlayerStatus(userID)
		status.Coins += resp.Reward.Coins
		_ = s.repo.SavePlayerStatus(status)
	}
	return resp, nil
}

func calcDamage(card model.CharacterCard, boss model.Boss) int {
	mult := attributeMultiplier(card.Element, boss.Element)
	base := float64(card.Attack) - float64(boss.Defense)/2 + float64(randInt(4))
	return max(1, int(math.Round(base*mult)))
}
func calcBossDamage(boss model.Boss, card model.CharacterCard) int {
	mult := attributeMultiplier(boss.Element, card.Element)
	base := float64(boss.Attack) - float64(card.Defense)/2 + float64(randInt(5))
	return max(1, int(math.Round(base*mult)))
}
func attributeMultiplier(attacker, defender string) float64 {
	strong := map[string]string{"fire": "wind", "wind": "earth", "earth": "water", "water": "fire", "light": "dark", "dark": "light"}
	weak := map[string]string{}
	for k, v := range strong {
		weak[v] = k
	}
	if strong[attacker] == defender {
		return 1.25
	}
	if weak[attacker] == defender {
		return 0.85
	}
	return 1.0
}
func (s *BossService) pickRewardCard(userID int64) (model.CharacterCard, bool) {
	candidates := s.repo.ListRewardCandidateCards()
	eligible := make([]model.CharacterCard, 0)
	for _, c := range candidates {
		if !s.repo.UserHasCard(userID, c.ID) {
			eligible = append(eligible, c)
		}
	}
	if len(eligible) == 0 {
		return model.CharacterCard{}, false
	}
	pick := eligible[randInt(len(eligible))]
	card, _, err := s.repo.AddCardToUser(userID, pick.ID)
	if err != nil {
		return model.CharacterCard{}, false
	}
	return card, true
}
func anyAlive(deck []battleCardState) bool {
	for _, c := range deck {
		if c.CurrentHP > 0 {
			return true
		}
	}
	return false
}
func randomAliveIndex(deck []battleCardState) int {
	alive := make([]int, 0)
	for i, c := range deck {
		if c.CurrentHP > 0 {
			alive = append(alive, i)
		}
	}
	if len(alive) == 0 {
		return -1
	}
	return alive[randInt(len(alive))]
}
func randInt(maxExclusive int) int {
	if maxExclusive <= 0 {
		return 0
	}
	n, err := crand.Int(crand.Reader, big.NewInt(int64(maxExclusive)))
	if err != nil {
		return 0
	}
	return int(n.Int64())
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
