package service

import (
	crand "crypto/rand"
	"fmt"
	"math"
	"math/big"
	"sort"

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

type bossAttackMove struct {
	Name       string
	EffectType string
	PowerBonus int
}

type defenseSkillResult struct {
	Name              string
	Tier              string
	EffectType        string
	DamageReduced     int
	HealAmount        int
	SupportTargetIdx  int
	SupportHealAmount int
	Revived           bool
	Evaded            bool
}

type bossAttackOutcome struct {
	Move              bossAttackMove
	TargetIdx         int
	Damage            int
	DefenseSkill      defenseSkillResult
	SupportTargetName string
	SupportTargetSlot int
	TargetKO          bool
}

const enduranceTargetTurns = 5

func (s *BossService) GetBoss(userID int64, bossID int64) (model.BossInfoResponse, error) {
	if bossID <= 0 {
		bossID = 1
	}
	boss, err := s.repo.GetBoss(bossID)
	if err != nil {
		return model.BossInfoResponse{}, err
	}
	boss = attachBossAttackMoves(boss)
	return model.BossInfoResponse{
		Boss:            boss,
		Bosses:          attachBossAttackMovesList(s.repo.ListBosses()),
		RecommendedDeck: s.recommendDeckForBoss(userID, boss),
		DropPreview:     s.bossDropPreview(boss),
	}, nil
}

func (s *BossService) AutoBattle(userID int64, bossID int64) (model.AutoBattleResponse, error) {
	if bossID <= 0 {
		bossID = 1
	}

	boss, err := s.repo.GetBoss(bossID)
	if err != nil {
		return model.AutoBattleResponse{}, err
	}

	deckView, err := s.repo.ListDeckCards(userID)
	if err != nil {
		return model.AutoBattleResponse{}, err
	}

	deck := buildBattleDeck(deckView)
	initialDeck := battleDeckSnapshot(deck)
	bossHP := boss.MaxHP

	survivedTurns, logs := runEnduranceBattle(boss, deck, bossHP)
	boss.CurrentHP = &bossHP

	resp := model.AutoBattleResponse{
		BattleTitle:   "ENDURANCE BATTLE",
		BattleMode:    "endurance",
		TargetTurns:   enduranceTargetTurns,
		SurvivedTurns: survivedTurns,
		Boss:          boss,
		InitialDeck:   initialDeck,
		Deck:          battleDeckSnapshot(deck),
		Logs:          logs,
	}

	if survivedTurns >= enduranceTargetTurns {
		s.applyEnduranceWinReward(userID, boss, &resp)
	} else {
		s.applyEnduranceLoseReward(userID, boss, &resp)
	}

	return resp, nil
}

func buildBattleDeck(deckView []model.DeckCardView) []battleCardState {
	deck := make([]battleCardState, 0, len(deckView))
	for _, dc := range deckView {
		deck = append(deck, battleCardState{
			Card:      dc.Card,
			DeckSlot:  dc.DeckSlot,
			CurrentHP: dc.Card.MaxHP,
		})
	}
	return deck
}

func battleDeckSnapshot(deck []battleCardState) []model.BattleDeckCardState {
	snapshot := make([]model.BattleDeckCardState, 0, len(deck))
	for _, dc := range deck {
		snapshot = append(snapshot, model.BattleDeckCardState{
			Card:       dc.Card,
			DeckSlot:   dc.DeckSlot,
			CurrentHP:  dc.CurrentHP,
			MaxHP:      dc.Card.MaxHP,
			KnockedOut: dc.CurrentHP <= 0,
		})
	}
	return snapshot
}

func runEnduranceBattle(boss model.Boss, deck []battleCardState, bossHP int) (int, []model.BattleLogEntry) {
	logs := make([]model.BattleLogEntry, 0, 64)
	survivedTurns := 0

	for round := 1; round <= enduranceTargetTurns && anyAlive(deck); round++ {
		logs = append(logs, enduranceTurnStartLog(round, boss, bossHP))

		for hit := 1; hit <= enduranceBossAttackCount(boss, round); hit++ {
			outcome, ok := resolveBossAttack(boss, deck, round, hit)
			if !ok {
				break
			}

			logs = append(logs, enduranceAttackLog(round, hit, boss, deck[outcome.TargetIdx], bossHP, outcome))
			if !anyAlive(deck) {
				break
			}
		}

		if !anyAlive(deck) {
			break
		}
		survivedTurns = round
	}

	return survivedTurns, logs
}

func enduranceTurnStartLog(round int, boss model.Boss, bossHP int) model.BattleLogEntry {
	return model.BattleLogEntry{
		Round:   round,
		Action:  "turn_start",
		Message: fmt.Sprintf("耐久 %d / %d ターン目。%s の猛攻に備えろ！", round, enduranceTargetTurns, boss.Name),
		BossHP:  bossHP,
	}
}

func resolveBossAttack(boss model.Boss, deck []battleCardState, round int, hit int) (bossAttackOutcome, bool) {
	targetIdx := randomAliveIndex(deck)
	if targetIdx < 0 {
		return bossAttackOutcome{}, false
	}

	move := selectBossAttackMove(boss, round, hit)
	rawDamage := calcEnduranceBossDamage(boss, deck[targetIdx].Card, round) + move.PowerBonus
	defenseSkill := triggerDefenseSkill(deck, targetIdx, boss, rawDamage)
	damage := max(0, rawDamage-defenseSkill.DamageReduced)

	applyDamageAndSelfRecovery(&deck[targetIdx], damage, &defenseSkill)
	supportTargetName, supportTargetSlot := applySupportRecovery(deck, &defenseSkill)

	return bossAttackOutcome{
		Move:              move,
		TargetIdx:         targetIdx,
		Damage:            damage,
		DefenseSkill:      defenseSkill,
		SupportTargetName: supportTargetName,
		SupportTargetSlot: supportTargetSlot,
		TargetKO:          deck[targetIdx].CurrentHP == 0,
	}, true
}

func applyDamageAndSelfRecovery(target *battleCardState, damage int, defenseSkill *defenseSkillResult) {
	target.CurrentHP = max(0, target.CurrentHP-damage)

	if defenseSkill.EffectType == "revive" && target.CurrentHP == 0 {
		target.CurrentHP = max(1, int(math.Round(float64(target.Card.MaxHP)*0.32)))
		defenseSkill.Revived = true
	}
	if defenseSkill.Tier == "advantage" && target.CurrentHP == 0 {
		target.CurrentHP = 1
		defenseSkill.Revived = true
	}
	if defenseSkill.HealAmount > 0 {
		target.CurrentHP = min(target.Card.MaxHP, target.CurrentHP+defenseSkill.HealAmount)
	}
}

func applySupportRecovery(deck []battleCardState, defenseSkill *defenseSkillResult) (string, int) {
	if defenseSkill.SupportTargetIdx < 0 ||
		defenseSkill.SupportTargetIdx >= len(deck) ||
		defenseSkill.SupportHealAmount <= 0 {
		return "", 0
	}

	supportTarget := &deck[defenseSkill.SupportTargetIdx]
	beforeHeal := supportTarget.CurrentHP
	supportTarget.CurrentHP = min(supportTarget.Card.MaxHP, supportTarget.CurrentHP+defenseSkill.SupportHealAmount)
	defenseSkill.SupportHealAmount = supportTarget.CurrentHP - beforeHeal
	if defenseSkill.SupportHealAmount <= 0 {
		return "", 0
	}

	return supportTarget.Card.Name, supportTarget.DeckSlot - 1
}

func enduranceAttackLog(round int, hit int, boss model.Boss, target battleCardState, bossHP int, outcome bossAttackOutcome) model.BattleLogEntry {
	defenseSkill := outcome.DefenseSkill

	return model.BattleLogEntry{
		Round:             round,
		ActorType:         "boss",
		ActorName:         boss.Name,
		TargetName:        target.Card.Name,
		Action:            "endurance_smash",
		SkillName:         outcome.Move.Name,
		EffectType:        outcome.Move.EffectType,
		DefenseSkillName:  defenseSkill.Name,
		DefenseSkillTier:  defenseSkill.Tier,
		DefenseEffectType: defenseSkill.EffectType,
		DamageReduced:     defenseSkill.DamageReduced,
		HealAmount:        defenseSkill.HealAmount,
		SupportTargetName: outcome.SupportTargetName,
		SupportTargetSlot: outcome.SupportTargetSlot,
		SupportHealAmount: defenseSkill.SupportHealAmount,
		Revived:           defenseSkill.Revived,
		Evaded:            defenseSkill.Evaded,
		Damage:            outcome.Damage,
		TargetHP:          target.CurrentHP,
		TargetMaxHP:       target.Card.MaxHP,
		TargetSlot:        target.DeckSlot - 1,
		CardHPAfter:       target.CurrentHP,
		BossHP:            bossHP,
		Message: enduranceBossAttackMessage(
			boss.Name,
			outcome.Move.Name,
			defenseSkill,
			round,
			hit,
			target.Card.Name,
			outcome.Damage,
			outcome.SupportTargetName,
		),
		TargetKO: outcome.TargetKO,
	}
}

func (s *BossService) applyEnduranceWinReward(userID int64, boss model.Boss, resp *model.AutoBattleResponse) {
	resp.Result = "win"
	resp.Reward.Exp = boss.RewardExp
	resp.Reward.Coins = boss.RewardCoins
	resp.Summary = fmt.Sprintf("%s の猛攻を %d ターン耐え切った！ EXP %d / COIN %d を獲得！", boss.Name, enduranceTargetTurns, boss.RewardExp, boss.RewardCoins)

	status, _ := s.repo.GetPlayerStatus(userID)
	status.Exp += boss.RewardExp
	status.Coins += boss.RewardCoins
	for status.Exp >= requiredExpForLevel(status.Level+1) {
		status.Level++
	}
	_ = s.repo.SavePlayerStatus(status)

	s.applyBossDropReward(userID, boss, resp)
}

func (s *BossService) applyBossDropReward(userID int64, boss model.Boss, resp *model.AutoBattleResponse) {
	dropRate := bossDropRate(boss)
	resp.Reward.BossDropRate = dropRate

	if randInt(100) >= dropRate {
		resp.Reward.BossDropMessage = fmt.Sprintf("ボスドロップ判定: 失敗（%d%%）", dropRate)
		return
	}

	reward, duplicate, ok := s.pickBossDropCard(userID, boss)
	if !ok {
		return
	}

	resp.Reward.BossDropCard = &reward
	resp.Reward.RewardCard = &reward
	resp.Reward.BossDropDuplicate = duplicate
	resp.Reward.BossDropOccurred = true
	if duplicate {
		resp.Reward.BossDropMessage = fmt.Sprintf("ボスドロップ: %s が重複し、カードが強化された。", reward.Name)
	} else {
		resp.Reward.BossDropMessage = fmt.Sprintf("ボスドロップ: %s を獲得！", reward.Name)
	}
	resp.Summary += " " + resp.Reward.BossDropMessage
}

func (s *BossService) applyEnduranceLoseReward(userID int64, boss model.Boss, resp *model.AutoBattleResponse) {
	resp.Result = "lose"
	resp.Reward.Coins = 20
	resp.Summary = fmt.Sprintf("%s の猛攻に %d / %d ターンで敗北... しかし COIN %d を持ち帰った。", boss.Name, resp.SurvivedTurns, enduranceTargetTurns, resp.Reward.Coins)

	status, _ := s.repo.GetPlayerStatus(userID)
	status.Coins += resp.Reward.Coins
	_ = s.repo.SavePlayerStatus(status)
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
func calcEnduranceBossDamage(boss model.Boss, card model.CharacterCard, round int) int {
	damage := calcBossDamage(boss, card)
	enrage := 1 + float64(max(0, round-1))*0.08
	return max(1, int(math.Round(float64(damage)*enrage)))
}
func triggerDefenseSkill(deck []battleCardState, targetIdx int, boss model.Boss, incomingDamage int) defenseSkillResult {
	card := deck[targetIdx]
	if attributeMultiplier(card.Card.Element, boss.Element) > 1 {
		advantageRate := 18 + card.Card.Rarity*2
		if randInt(100) < advantageRate {
			skill := card.Card.AdvantageDefenseSkill
			if skill.Name == "" {
				skill = defaultAdvantageDefenseSkill(card.Card)
			}
			reduced := max(1, int(math.Round(float64(incomingDamage)*0.72)))
			heal := max(2, int(math.Round(float64(card.Card.MaxHP)*0.18)))
			return defenseSkillResult{Name: skill.Name, Tier: "advantage", EffectType: "advantage", DamageReduced: reduced, HealAmount: heal, SupportTargetIdx: -1}
		}
	}
	uniqueRate := card.Card.UniqueSkill.TriggerRate
	if uniqueRate <= 0 {
		uniqueRate = 30 + card.Card.Rarity*3
	}
	if randInt(100) < uniqueRate {
		skill := card.Card.UniqueSkill
		if skill.Name == "" {
			skill = defaultUniqueDefenseSkill(card.Card)
		}
		return applyUniqueSkillEffect(skill, deck, targetIdx, incomingDamage)
	}
	return defenseSkillResult{}
}

func applyUniqueSkillEffect(skill model.DefenseSkill, deck []battleCardState, targetIdx int, incomingDamage int) defenseSkillResult {
	effectType := skill.EffectType
	if effectType == "" {
		effectType = "mitigate"
	}
	result := defenseSkillResult{Name: skill.Name, Tier: "unique", EffectType: effectType, SupportTargetIdx: -1}
	switch effectType {
	case "shield":
		result.DamageReduced = max(1, int(math.Round(float64(incomingDamage)*0.58)))
	case "heal":
		result.DamageReduced = max(1, int(math.Round(float64(incomingDamage)*0.24)))
		result.SupportTargetIdx = lowestAliveHPIndex(deck, targetIdx)
		if result.SupportTargetIdx >= 0 {
			result.SupportHealAmount = max(2, int(math.Round(float64(deck[result.SupportTargetIdx].Card.MaxHP)*0.16)))
		}
	case "evade":
		result.DamageReduced = incomingDamage
		result.Evaded = true
	case "revive":
		result.DamageReduced = max(1, int(math.Round(float64(incomingDamage)*0.18)))
	default:
		result.EffectType = "mitigate"
		result.DamageReduced = max(1, int(math.Round(float64(incomingDamage)*0.42)))
	}
	return result
}

func defaultUniqueDefenseSkill(card model.CharacterCard) model.DefenseSkill {
	return model.DefenseSkill{
		Name:        card.Name + "の固有防御",
		Description: "カード固有の発生しやすい防御。被弾時にダメージを軽減します。",
		TriggerRate: 30 + card.Rarity*3,
		EffectType:  "mitigate",
	}
}
func defaultAdvantageDefenseSkill(card model.CharacterCard) model.DefenseSkill {
	return model.DefenseSkill{
		Name:        card.Name + "の特攻防御",
		Description: "攻撃相性で有利なボスから攻撃を受ける時だけ一定確率で発生します。",
		TriggerRate: 18 + card.Rarity*2,
		EffectType:  "advantage",
	}
}
func lowestAliveHPIndex(deck []battleCardState, fallbackIdx int) int {
	bestIdx := -1
	bestPct := math.MaxFloat64
	for i, card := range deck {
		if card.CurrentHP <= 0 {
			continue
		}
		pct := float64(card.CurrentHP) / float64(max(1, card.Card.MaxHP))
		if pct < bestPct {
			bestPct = pct
			bestIdx = i
		}
	}
	if bestIdx >= 0 {
		return bestIdx
	}
	if fallbackIdx >= 0 && fallbackIdx < len(deck) {
		return fallbackIdx
	}
	return -1
}
func enduranceBossAttackCount(boss model.Boss, round int) int {
	switch {
	case boss.ID >= 6 && round >= 2:
		return 2
	case boss.ID >= 5 && round >= 3:
		return 2
	case boss.ID >= 3 && round >= 4:
		return 2
	default:
		return 1
	}
}
func selectBossAttackMove(boss model.Boss, round int, hit int) bossAttackMove {
	moves := bossAttackMoves()[boss.ID]
	if len(moves) == 0 {
		moves = elementalBossMoves(boss.Element)
	}
	index := (round + hit + int(boss.ID)) % len(moves)
	return moves[index]
}
func attachBossAttackMovesList(bosses []model.Boss) []model.Boss {
	out := make([]model.Boss, 0, len(bosses))
	for _, boss := range bosses {
		out = append(out, attachBossAttackMoves(boss))
	}
	return out
}
func attachBossAttackMoves(boss model.Boss) model.Boss {
	moves := bossAttackMoves()[boss.ID]
	if len(moves) == 0 {
		moves = elementalBossMoves(boss.Element)
	}
	boss.AttackMoves = make([]model.BossAttackMove, 0, len(moves))
	for _, move := range moves {
		boss.AttackMoves = append(boss.AttackMoves, model.BossAttackMove{Name: move.Name, EffectType: move.EffectType, PowerBonus: move.PowerBonus})
	}
	boss.StrategyHint = buildBossStrategyHint(boss, moves)
	return boss
}

func buildBossStrategyHint(boss model.Boss, moves []bossAttackMove) model.BossStrategyHint {
	return model.BossStrategyHint{
		EffectiveElements: effectivePlayerElements(boss, moves),
		DangerousMoves:    dangerousMoveHints(moves),
		RecommendedCards:  recommendedCardTrends(boss, moves),
	}
}

func effectivePlayerElements(boss model.Boss, moves []bossAttackMove) []string {
	switch boss.Element {
	case "fire", "earth":
		return []string{"body", "heart"}
	case "water", "wind":
		return []string{"tech", "body"}
	case "light", "dark":
		return []string{"heart", "tech"}
	}
	counts := map[string]int{"heart": 0, "tech": 0, "body": 0}
	for _, move := range moves {
		for _, element := range elementsForEffect(move.EffectType) {
			counts[element] += 1 + move.PowerBonus
		}
	}
	return topElements(counts, 2)
}

func elementsForEffect(effect string) []string {
	switch effect {
	case "abyss", "void", "holy", "cosmic":
		return []string{"heart"}
	case "storm", "water", "ice", "spike":
		return []string{"tech"}
	case "fire", "quake", "tentacle", "fang", "venom":
		return []string{"body"}
	default:
		return []string{"body"}
	}
}

func topElements(counts map[string]int, limit int) []string {
	order := []string{"heart", "tech", "body"}
	for i := 0; i < len(order); i++ {
		for j := i + 1; j < len(order); j++ {
			if counts[order[j]] > counts[order[i]] {
				order[i], order[j] = order[j], order[i]
			}
		}
	}
	if limit > len(order) {
		limit = len(order)
	}
	return order[:limit]
}

func dangerousMoveHints(moves []bossAttackMove) []string {
	if len(moves) == 0 {
		return nil
	}
	strongest := moves[0]
	for _, move := range moves[1:] {
		if move.PowerBonus > strongest.PowerBonus {
			strongest = move
		}
	}
	hints := []string{fmt.Sprintf("%s: 追加威力 +%d の主砲。HP5割未満のカードに集中すると危険。", strongest.Name, strongest.PowerBonus)}
	for _, move := range moves {
		if move.Name == strongest.Name {
			continue
		}
		switch move.EffectType {
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
	return hints
}

func recommendedCardTrends(boss model.Boss, moves []bossAttackMove) []string {
	counts := map[string]int{}
	for _, move := range moves {
		counts[move.EffectType]++
	}
	recommendations := []string{}
	if counts["fire"] > 0 || counts["quake"] > 0 || counts["tentacle"] > 0 {
		recommendations = appendUniqueString(recommendations, "体属性の盾役・軽減カードを前線に置く")
	}
	if counts["water"] > 0 || counts["venom"] > 0 || boss.Element == "water" {
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
	for _, existing := range items {
		if existing == item {
			return items
		}
	}
	return append(items, item)
}
func bossAttackMoves() map[int64][]bossAttackMove {
	return map[int64][]bossAttackMove{
		1:  {{"沈石の触腕", "quake", 0}, {"偶像圧砕", "tentacle", 1}, {"海底神殿の重圧", "abyss", 2}},
		2:  {{"深淵潮流", "abyss", 0}, {"ダゴンの咆哮", "tentacle", 1}, {"鱗波の大津波", "water", 2}},
		3:  {{"星火触腕", "fire", 1}, {"クトゥグアの火環", "fire", 2}, {"焼尽する赤星", "cosmic", 2}},
		4:  {{"落とし子の翼撃", "storm", 0}, {"蛸面旋風", "tentacle", 1}, {"星風滑翔", "storm", 2}},
		5:  {{"千眼光輪", "holy", 1}, {"白金翼の審判", "holy", 1}, {"旧支配者の凝視", "cosmic", 2}},
		6:  {{"黒いファラオの囁き", "void", 1}, {"夢裂きの触腕", "tentacle", 1}, {"這い寄る混沌", "abyss", 2}},
		7:  {{"不定形圧潰", "tentacle", 0}, {"警報眼の乱反射", "cosmic", 1}, {"ショゴス濁流", "abyss", 2}},
		8:  {{"黄昏急降下", "storm", 0}, {"星騎の鉤爪", "fang", 1}, {"腐翼の乱気流", "storm", 2}},
		9:  {{"深海詠唱", "water", 0}, {"魚鱗の祈祷刃", "spike", 1}, {"司祭の水圧牢", "abyss", 2}},
		10: {{"時間差光撃", "holy", 0}, {"イスの記録光線", "cosmic", 1}, {"時渡りの断章", "void", 2}},
		11: {{"石化凝視", "quake", 0}, {"火山島の仮面", "fire", 1}, {"ガタノトーアの鈍光", "abyss", 2}},
		12: {{"菌糸外科", "venom", 0}, {"ミ＝ゴの摘出針", "spike", 1}, {"脳髄標本の閃き", "cosmic", 2}},
		13: {{"黄衣の幕開け", "void", 1}, {"ハスターの名状舞台", "abyss", 1}, {"王の仮面光", "holy", 2}},
		14: {{"星蜘蛛の粘糸", "spike", 0}, {"アトラックの大橋", "quake", 1}, {"奈落へ編む糸", "void", 2}},
		15: {{"白嵐の踏破", "ice", 0}, {"イタカの凍息", "ice", 1}, {"極北の裂風", "storm", 2}},
		16: {{"蛇父の毒牙", "venom", 0}, {"黄金眼の威圧", "quake", 1}, {"イグの巻き締め", "fang", 2}},
		17: {{"黒睡の呪吐", "void", 0}, {"蟇王の鈍撃", "quake", 1}, {"ツァトゥグァの眠泥", "abyss", 2}},
		18: {{"極圏冷線", "ice", 0}, {"古の五放射器官", "cosmic", 1}, {"氷棘標本", "spike", 2}},
		19: {{"無貌の夜襲", "void", 1}, {"夜鬼の羽音", "storm", 1}, {"ナイトゴーント急襲", "fang", 2}},
		20: {{"不可視の吸血", "void", 1}, {"星血の赤光", "cosmic", 1}, {"透明な捕食爪", "fang", 2}},
		21: {{"鋭角跳躍", "storm", 0}, {"ティンダロスの牙", "fang", 1}, {"時間角の裂傷", "void", 2}},
		22: {{"精神圧波", "cosmic", 0}, {"ロイガーの念竜巻", "storm", 1}, {"光る思念杭", "holy", 2}},
		23: {{"湖底棘柱", "spike", 0}, {"グラーキの水晶波", "water", 1}, {"濁湖の背骨", "abyss", 2}},
		24: {{"象牙吸血", "fang", 0}, {"石像神の踏鳴", "quake", 1}, {"チャウグナーの鼻撃", "tentacle", 2}},
		25: {{"凍爪跳躍", "ice", 0}, {"ラーン＝テゴスの飢牙", "fang", 1}, {"氷河の昆虫脚", "spike", 2}},
		26: {{"屍都の黒炎", "void", 1}, {"モルディギアンの墓霧", "abyss", 1}, {"地下王の葬列", "quake", 2}},
		27: {{"双頭噛砕", "fang", 1}, {"ズシャーの星熱波", "fire", 2}, {"二重螺旋火花", "cosmic", 2}},
		28: {{"星司祭の深祈", "water", 1}, {"沈没神殿の鐘", "abyss", 1}, {"クトゥルフ聖印", "tentacle", 2}},
		29: {{"無秩序核熱", "fire", 2}, {"アザトースの不協音", "cosmic", 2}, {"盲目の星爆", "void", 3}},
		30: {{"門の開放", "cosmic", 2}, {"ヨグの球体連鎖", "holy", 2}, {"全ての時空圧", "void", 3}},
	}
}
func elementalBossMoves(element string) []bossAttackMove {
	switch element {
	case "fire":
		return []bossAttackMove{{"災火の一撃", "fire", 1}, {"焦熱波", "fire", 1}, {"赤星の残響", "cosmic", 2}}
	case "water":
		return []bossAttackMove{{"深水圧", "water", 0}, {"濁流撃", "abyss", 1}, {"水棘波", "spike", 2}}
	case "wind":
		return []bossAttackMove{{"裂風爪", "storm", 0}, {"乱気流", "storm", 1}, {"急襲牙", "fang", 2}}
	case "light":
		return []bossAttackMove{{"光輪撃", "holy", 0}, {"星辰光", "cosmic", 1}, {"裁きの閃光", "holy", 2}}
	case "dark":
		return []bossAttackMove{{"闇裂き", "void", 0}, {"深淵波", "abyss", 1}, {"黒い触腕", "tentacle", 2}}
	default:
		return []bossAttackMove{{"重圧撃", "quake", 0}, {"異形の一撃", "tentacle", 1}, {"レリック崩壊波", "cosmic", 2}}
	}
}
func enduranceBossAttackMessage(bossName string, skillName string, defenseSkill defenseSkillResult, round int, hit int, targetName string, damage int, supportTargetName string) string {
	guardText := ""
	if defenseSkill.Name != "" {
		if defenseSkill.Tier == "advantage" {
			guardText = fmt.Sprintf(" %s が特攻防御「%s」を発動、%d軽減・HP%d回復！", targetName, defenseSkill.Name, defenseSkill.DamageReduced, defenseSkill.HealAmount)
		} else {
			guardText = fmt.Sprintf(" %s が固有スキル「%s」を発動、%d軽減！", targetName, defenseSkill.Name, defenseSkill.DamageReduced)
			switch defenseSkill.EffectType {
			case "shield":
				guardText = fmt.Sprintf(" %s が盾役「%s」で攻撃を受け止め、%d軽減！", targetName, defenseSkill.Name, defenseSkill.DamageReduced)
			case "heal":
				if supportTargetName != "" && defenseSkill.SupportHealAmount > 0 {
					guardText = fmt.Sprintf(" %s が回復補助「%s」を発動、%d軽減・%sをHP%d回復！", targetName, defenseSkill.Name, defenseSkill.DamageReduced, supportTargetName, defenseSkill.SupportHealAmount)
				} else {
					guardText = fmt.Sprintf(" %s が回復補助「%s」を発動、%d軽減！", targetName, defenseSkill.Name, defenseSkill.DamageReduced)
				}
			case "evade":
				guardText = fmt.Sprintf(" %s が回避「%s」で攻撃を無効化！", targetName, defenseSkill.Name)
			case "revive":
				if defenseSkill.Revived {
					guardText = fmt.Sprintf(" %s が蘇生「%s」で戦線復帰！", targetName, defenseSkill.Name)
				} else {
					guardText = fmt.Sprintf(" %s が蘇生準備「%s」で%d軽減！", targetName, defenseSkill.Name, defenseSkill.DamageReduced)
				}
			}
		}
	}
	if hit <= 1 {
		return fmt.Sprintf("%s の「%s」！ %s に %d ダメージ！%s", bossName, skillName, targetName, damage, guardText)
	}
	return fmt.Sprintf("%s の追撃「%s」！ %s に %d ダメージ！%s", bossName, skillName, targetName, damage, guardText)
}
func attributeMultiplier(attacker, defender string) float64 {
	strong := map[string]string{"heart": "tech", "tech": "body", "body": "heart", "fire": "wind", "wind": "earth", "earth": "water", "water": "fire", "light": "dark", "dark": "light"}
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
func (s *BossService) bossDropPreview(boss model.Boss) model.BossDropPreview {
	return model.BossDropPreview{DropRatePercent: bossDropRate(boss), Candidates: bossDropCandidatesFromPool(boss, s.repo.ListRewardCandidateCards())}
}

func bossDropRate(boss model.Boss) int {
	rate := 22 + int(boss.ID%5)*2
	if boss.ID >= 20 {
		rate += 4
	}
	if boss.ID >= 29 {
		rate += 6
	}
	return min(42, rate)
}

func (s *BossService) pickBossDropCard(userID int64, boss model.Boss) (model.CharacterCard, bool, bool) {
	candidates := bossDropCandidatesFromPool(boss, s.repo.ListRewardCandidateCards())
	if len(candidates) == 0 {
		return model.CharacterCard{}, false, false
	}
	unowned := make([]model.CharacterCard, 0, len(candidates))
	for _, card := range candidates {
		if !s.repo.UserHasCard(userID, card.ID) {
			unowned = append(unowned, card)
		}
	}
	pool := candidates
	if len(unowned) > 0 {
		pool = unowned
	}
	pick := pool[randInt(len(pool))]
	card, duplicate, err := s.repo.AddCardToUser(userID, pick.ID)
	if err != nil {
		return model.CharacterCard{}, false, false
	}
	return card, duplicate, true
}

func bossDropCandidatesFromPool(boss model.Boss, pool []model.CharacterCard) []model.CharacterCard {
	if len(pool) == 0 {
		return nil
	}
	effective := effectivePlayerElements(boss, bossAttackMoves()[boss.ID])
	effectiveSet := map[string]bool{}
	for _, element := range effective {
		effectiveSet[element] = true
	}
	candidates := make([]model.CharacterCard, 0, 5)
	for _, card := range pool {
		if effectiveSet[card.Element] {
			candidates = append(candidates, card)
		}
	}
	if len(candidates) == 0 {
		candidates = append(candidates, pool...)
	}
	sort.Slice(candidates, func(i, j int) bool {
		left := int((candidates[i].ID + boss.ID*7) % 97)
		right := int((candidates[j].ID + boss.ID*7) % 97)
		if left == right {
			return candidates[i].Rarity > candidates[j].Rarity
		}
		return left < right
	})
	if len(candidates) > 4 {
		candidates = candidates[:4]
	}
	return candidates
}

func (s *BossService) recommendDeckForBoss(userID int64, boss model.Boss) []model.BossRecommendedCard {
	owned, err := s.repo.ListOwnedCards(userID)
	if err != nil {
		return nil
	}
	moves := boss.AttackMoves
	if len(moves) == 0 {
		boss = attachBossAttackMoves(boss)
		moves = boss.AttackMoves
	}
	effective := map[string]bool{}
	for _, element := range boss.StrategyHint.EffectiveElements {
		effective[element] = true
	}
	if len(effective) == 0 {
		for _, element := range effectivePlayerElements(boss, bossAttackMoves()[boss.ID]) {
			effective[element] = true
		}
	}
	out := make([]model.BossRecommendedCard, 0, len(owned))
	for _, entry := range owned {
		score, reason := recommendationScore(entry, effective, moves)
		out = append(out, model.BossRecommendedCard{Card: entry.Card, Reason: reason, Score: score, Owned: true, InDeck: entry.InDeck, DeckSlot: entry.User.DeckSlot})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			if out[i].Card.Rarity == out[j].Card.Rarity {
				return out[i].Card.ID < out[j].Card.ID
			}
			return out[i].Card.Rarity > out[j].Card.Rarity
		}
		return out[i].Score > out[j].Score
	})
	if len(out) > 6 {
		out = out[:6]
	}
	return out
}

func recommendationScore(entry model.CardCollectionEntry, effective map[string]bool, moves []model.BossAttackMove) (int, string) {
	card := entry.Card
	score := card.MaxHP + card.Defense*4 + card.Rarity*8
	reasons := make([]string, 0, 3)
	if effective[card.Element] {
		score += 40
		reasons = append(reasons, "有効属性")
	}
	effect := card.UniqueSkill.EffectType
	switch effect {
	case "shield":
		score += 34
		reasons = append(reasons, "盾役")
	case "heal":
		score += 30
		reasons = append(reasons, "回復補助")
	case "evade":
		score += 28
		reasons = append(reasons, "回避")
	case "revive":
		score += 32
		reasons = append(reasons, "蘇生")
	case "mitigate":
		score += 24
		reasons = append(reasons, "軽減")
	}
	if bossHasHeavyPhysicalMoves(moves) && (effect == "shield" || effect == "mitigate" || card.Element == "body") {
		score += 18
		reasons = append(reasons, "重圧対策")
	}
	if bossHasMentalMoves(moves) && (effect == "evade" || effect == "revive" || card.Element == "heart") {
		score += 18
		reasons = append(reasons, "深淵対策")
	}
	if bossHasWaveMoves(moves) && effect == "heal" {
		score += 18
		reasons = append(reasons, "波紋対策")
	}
	if entry.InDeck {
		score += 8
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "耐久補助")
	}
	return score, joinReasons(reasons)
}

func bossHasHeavyPhysicalMoves(moves []model.BossAttackMove) bool {
	for _, move := range moves {
		if move.EffectType == "fire" || move.EffectType == "quake" || move.EffectType == "tentacle" || move.EffectType == "fang" {
			return true
		}
	}
	return false
}

func bossHasMentalMoves(moves []model.BossAttackMove) bool {
	for _, move := range moves {
		if move.EffectType == "abyss" || move.EffectType == "void" || move.EffectType == "cosmic" {
			return true
		}
	}
	return false
}

func bossHasWaveMoves(moves []model.BossAttackMove) bool {
	for _, move := range moves {
		if move.EffectType == "water" || move.EffectType == "venom" || move.EffectType == "ice" {
			return true
		}
	}
	return false
}

func joinReasons(reasons []string) string {
	seen := map[string]bool{}
	out := ""
	for _, reason := range reasons {
		if seen[reason] {
			continue
		}
		seen[reason] = true
		if out != "" {
			out += " / "
		}
		out += reason
	}
	return out
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
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
