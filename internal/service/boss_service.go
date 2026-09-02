package service

import (
	crand "crypto/rand"
	"fmt"
	"math/big"

	"retro-treasure-backend/internal/model"
	"retro-treasure-backend/internal/repository"
	"slices"
)

type BossService struct{ repo *repository.MemoryRepository }

func NewBossService(repo *repository.MemoryRepository) *BossService {
	return &BossService{repo: repo}
}

type battleCardState struct {
	Card          model.CharacterCard
	DeckSlot      int
	CurrentHP     int
	CurrentShield int
	CurrentEvade  bool
	SkillUsed     bool
}

func (s *BossService) GetBoss(userID int64) ([]model.BossInfoResponse, error) {
	status, err := s.repo.GetPlayerStatus(userID)
	if err != nil {
		return []model.BossInfoResponse{}, err
	}

	owned, err := s.repo.ListOwnedCards(userID)
	if err != nil {
		return []model.BossInfoResponse{}, err
	}

	bosses := []model.BossInfoResponse{}
	for _, boss := range s.repo.ListBosses() {
		if status.Tutorial && len(owned) == 2 && boss.ID != 1 {
			continue
		}
		if status.Tutorial && len(owned) == 3 && boss.ID != 2 {
			continue
		}
		item := model.BossInfoResponse{
			Boss:         boss,
			StrategyHint: buildBossStrategyHint(boss),
			Drops:        []model.BossDrop{{DropRate: boss.DropRate, Candidates: []model.BossDropItem{}}},
		}
		for _, card := range s.repo.ListCards(boss.Candidates) {
			exists := false
			for _, hand := range owned {
				if card.ID == hand.Card.ID {
					exists = true
					break
				}
			}
			item.Drops[0].Candidates = append(item.Drops[0].Candidates, model.BossDropItem{
				Owned: exists,
				Card:  card,
			})
		}
		bosses = append(bosses, item)
	}

	return bosses, nil
}

func (s *BossService) AutoBattle(userID int64, bossID int64) (model.AutoBattleResponse, error) {
	boss, err := s.repo.GetBoss(bossID)
	if err != nil {
		return model.AutoBattleResponse{}, err
	}

	deckView, err := s.repo.ListDeckCards(userID)
	if err != nil {
		return model.AutoBattleResponse{}, err
	}

	deck := []battleCardState{}
	for _, dc := range deckView {
		deck = append(deck, battleCardState{
			Card:          dc.Card,
			DeckSlot:      dc.DeckSlot,
			CurrentHP:     dc.Card.MaxHP,
			CurrentShield: 0,
			CurrentEvade:  false,
			SkillUsed:     false,
		})
	}
	survivedTurns, logs := runEnduranceBattle(boss, deck)

	deckMapped := map[int]model.CharacterCard{}
	for _, card := range deckView {
		deckMapped[card.DeckSlot] = card.Card
	}

	resp := model.AutoBattleResponse{
		Boss:        boss,
		InitialDeck: deckMapped,
		Logs:        logs,
		IsWin:       survivedTurns >= len(boss.AttackMoves),
	}

	return resp, nil
}

func runEnduranceBattle(boss model.Boss, deck []battleCardState) (int, []model.BattleLogEntry) {
	logs := []model.BattleLogEntry{}
	survivedTurns := 0

	// ラウンド
	for round, moves := range boss.AttackMoves {
		// ターン
		logs = append(logs, model.BattleLogEntry{
			Round:     round,
			ActorType: "turn_start",
		})

		// 攻撃
		for skill_index, move := range moves {
			indexes := randomAliveIndex(deck)
			if move.All {
				indexes = allAliveIndex(deck)
			}

			// スキル発動（先制）
			logs = append(logs, triggerDefenseSkill(deck, indexes, move, true, round)...)

			for i, index := range indexes {
				target := &deck[index]
				if target.CurrentHP <= 0 {
					continue
				}

				// ボスの攻撃
				damage := move.Power
				if target.CurrentEvade {
					damage = 0
					target.CurrentEvade = false
				} else if target.CurrentShield > 0 {
					shield := min(damage, target.CurrentShield)
					damage = max(0, damage-shield)
					target.CurrentShield = max(0, target.CurrentShield-shield)
				}
				defence := 0
				if move.Element == target.Card.Element {
					defence = target.Card.Defense
				}
				damage = max(0, damage-defence)
				target.CurrentHP = max(0, target.CurrentHP-damage)

				// ボスの攻撃ログ
				log := model.BattleLogEntry{
					Round:           round,
					SkillIndex:      skill_index,
					ActorType:       "boss",
					Damage:          damage,
					TargetSlot:      target.DeckSlot,
					CardHPAfter:     target.CurrentHP,
					CardShieldAfter: target.CurrentShield,
					CardEvadeAfter:  target.CurrentEvade,
					Index:           []int{},
				}
				if i == 0 {
					log.Index = indexes
				}
				logs = append(logs, log)
			}
			if !anyAlive(deck) {
				break
			}

			// スキル発動（後攻）
			logs = append(logs, triggerDefenseSkill(deck, indexes, move, false, round)...)
		}
		if !anyAlive(deck) {
			break
		}
		survivedTurns = round + 1
	}

	return survivedTurns, logs
}

func (s *BossService) GetReward(userID int64, bossID int64) (model.AutoResultResponse, error) {
	boss, err := s.repo.GetBoss(bossID)
	if err != nil {
		return model.AutoResultResponse{}, err
	}

	resp := model.AutoResultResponse{
		Exp:        boss.RewardExp,
		Coins:      boss.RewardCoins,
		Dropped:    false,
		Duplicate:  false,
		DropDice:   randInt(100),
		RewardCard: nil,
	}

	if resp.DropDice < boss.DropRate {
		resp.Dropped = true
		reward, ok := s.pickBossDropCard(userID, boss) // ここにカードの追加処理が入っている
		if ok {
			resp.RewardCard = &reward
		} else {
			resp.Duplicate = true
			resp.Coins += 50
		}
	}

	status, _ := s.repo.GetPlayerStatus(userID)
	status.Exp += boss.RewardExp
	status.Coins += boss.RewardCoins
	// TODO: ここにレベルアップ処理
	// for status.Exp >= requiredExpForLevel(status.Level+1) {
	// 	status.Level++
	// }
	_ = s.repo.SavePlayerStatus(status)

	return resp, nil
}

func triggerDefenseSkill(deck []battleCardState, targetIndexes []int, move model.BossAttackMove, first bool, round int) []model.BattleLogEntry {
	result := []model.BattleLogEntry{}
	for userIndex := range deck {
		user := &deck[userIndex]
		if user.SkillUsed || user.CurrentHP <= 0 {
			continue
		}
		for skill_index, skill := range user.Card.Skills {
			if skill.IsFirst != first {
				continue
			}
			if skill.SubElement != "" && skill.SubElement != move.SubElement {
				continue
			}
			if randInt(100) < skill.TriggerRate {
				supportIndexes := []int{}
				if skill.Target == "target" {
					supportIndexes = append(supportIndexes, targetIndexes[randInt(len(targetIndexes))])
				}
				if skill.Target == "all" {
					for i := range deck {
						supportIndexes = append(supportIndexes, i)
					}
				}
				used := false
				for i, supportIndex := range supportIndexes {
					support := &deck[supportIndex]
					if skill.IsDamaged && (support.CurrentHP == 0 || support.CurrentHP >= support.Card.MaxHP) {
						continue
					}
					if skill.IsDead && support.CurrentHP > 0 {
						continue
					}
					effected := false
					if skill.Heal > 0 && support.CurrentHP < support.Card.MaxHP {
						support.CurrentHP = min(support.Card.MaxHP, support.CurrentHP+skill.Heal)
						effected = true
					}
					if skill.Shield > support.CurrentShield && support.CurrentHP > 0 {
						support.CurrentShield = skill.Shield
						effected = true
					}
					if skill.IsEvade {
						support.CurrentEvade = true
						effected = true
					}
					if effected {
						log := model.BattleLogEntry{
							Round:           round,
							SkillIndex:      skill_index,
							ActorType:       "skill",
							ActorSlot:       user.DeckSlot,
							TargetSlot:      support.DeckSlot,
							Damage:          skill.Heal,
							CardHPAfter:     support.CurrentHP,
							CardShieldAfter: support.CurrentShield,
							CardEvadeAfter:  support.CurrentEvade,
							Index:           []int{},
						}
						if i == 0 {
							log.Index = supportIndexes
						}
						result = append(result, log)
						used = true
					}
				}
				if used {
					user.SkillUsed = true
					break
				}
			}
		}
	}
	return result
}

func buildBossStrategyHint(boss model.Boss) model.BossStrategyHint {
	return model.BossStrategyHint{
		EffectiveElements: effectivePlayerElements(boss),
		DangerousMoves:    dangerousMoveHints(boss),
		RecommendedCards:  recommendedCardTrends(boss),
	}
}

func effectivePlayerElements(boss model.Boss) []string {
	counts := map[string]int{}
	for _, moves := range boss.AttackMoves {
		for _, move := range moves {
			counts[move.Element]++
			if len(move.SubElement) > 0 {
				counts[move.SubElement]++
			}
		}
	}
	return topElements(counts)
}

func topElements(counts map[string]int) []string {
	order := []string{}
	for key := range counts {
		order = append(order, key)
	}
	for i := range order {
		for j := i + 1; j < len(order); j++ {
			if counts[order[j]] > counts[order[i]] {
				order[i], order[j] = order[j], order[i]
			}
		}
	}
	return order
}

func dangerousMoveHints(boss model.Boss) []string {
	var strongest *model.BossAttackMove
	for _, moves := range boss.AttackMoves {
		for _, move := range moves {
			if strongest == nil || move.Power > strongest.Power {
				strongest = &move
			}
		}
	}
	hints := []string{fmt.Sprintf("%s: 威力 %d の主砲。HP5割未満のカードに集中すると危険。", strongest.Name, strongest.Power)}
	for _, moves := range boss.AttackMoves {
		for _, move := range moves {
			if move.Name == strongest.Name {
				continue
			}
			switch move.SubElement {
			case "fire":
				hints = append(hints, move.Name+": 爆発系。低HPカードの連続被弾に注意。")
			case "water":
				hints = append(hints, move.Name+": 波紋系。回復補助で立て直したい。")
			case "abyss", "void":
				hints = append(hints, move.Name+": 精神圧系。心属性の回避・蘇生が有効。")
			case "tentacle", "quake":
				hints = append(hints, move.Name+": 拘束/重圧系。盾役で受けたい。")
			case "ice":
				hints = append(hints, move.Name+": 凍結系。蘇生役を温存したい。")
			}
			if len(hints) >= 2 {
				break
			}
		}
	}
	return hints
}

func recommendedCardTrends(boss model.Boss) []string {
	counts := map[string]int{}
	for _, moves := range boss.AttackMoves {
		for _, move := range moves {
			counts[move.Element]++
			if len(move.SubElement) > 0 {
				counts[move.SubElement]++
			}
		}
	}
	recommendations := []string{}
	if counts["fire"] > 0 || counts["quake"] > 0 || counts["tentacle"] > 0 {
		recommendations = appendUniqueString(recommendations, "体属性の盾役・軽減カードを前線に置く")
	}
	if counts["water"] > 0 || counts["venom"] > 0 {
		recommendations = appendUniqueString(recommendations, "心属性の回復補助カードでHPを戻す")
	}
	if counts["abyss"] > 0 || counts["void"] > 0 || counts["cosmic"] > 0 {
		recommendations = appendUniqueString(recommendations, "心属性の回避・蘇生カードを混ぜる")
	}
	if counts["storm"] > 0 || counts["spike"] > 0 || counts["ice"] > 0 {
		recommendations = appendUniqueString(recommendations, "技属性の回避カードで連撃をかわす")
	}
	if len(recommendations) == 0 {
		recommendations = appendUniqueString(recommendations, "盾役・回復補助・蘇生を1枚ずつ入れる")
	}
	fallbacks := []string{
		"HP5割未満を赤信号として、軽減スキル持ちを優先する",
		"蘇生カードを1枚入れて終盤の事故に備える",
		"回復補助と盾役を同時に採用して耐久ターンを伸ばす",
	}
	for _, fallback := range fallbacks {
		if len(recommendations) >= 3 {
			break
		}
		recommendations = appendUniqueString(recommendations, fallback)
	}
	return recommendations[:3]
}

func appendUniqueString(items []string, item string) []string {
	if slices.Contains(items, item) {
		return items
	}
	return append(items, item)
}

func (s *BossService) pickBossDropCard(userID int64, boss model.Boss) (model.CharacterCard, bool) {
	if len(boss.Candidates) == 0 {
		return model.CharacterCard{}, false
	}
	unowned := []int64{} // 未所持優先の仕組みがあるらしい
	for _, cardId := range boss.Candidates {
		if !s.repo.UserHasCard(userID, cardId) {
			unowned = append(unowned, cardId)
		}
	}
	if len(unowned) == 0 {
		return model.CharacterCard{}, false
	}
	pick := unowned[randInt(len(unowned))]
	card, duplicate, err := s.repo.AddCardToUser(userID, pick)
	if duplicate || err != nil {
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
func randomAliveIndex(deck []battleCardState) []int {
	alive := allAliveIndex(deck)
	if len(alive) == 0 {
		return []int{}
	}
	return []int{alive[randInt(len(alive))]}
}
func allAliveIndex(deck []battleCardState) []int {
	alive := make([]int, 0)
	for i, c := range deck {
		if c.CurrentHP > 0 {
			alive = append(alive, i)
		}
	}
	return alive
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
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
