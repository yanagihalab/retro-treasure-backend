package repository

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"retro-treasure-backend/internal/model"
)

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrAreaNotFound        = errors.New("area not found")
	ErrInsufficientStamina = errors.New("insufficient stamina")
	ErrAreaLocked          = errors.New("area locked")
	ErrAlreadyClaimed      = errors.New("login bonus already claimed today")
	ErrCardNotFound        = errors.New("card not found")
	ErrBossNotFound        = errors.New("boss not found")
)

type WeightedDrop struct {
	ItemID     *int64
	ResultType string
	Weight     int
	ExpReward  int
	CoinReward int
	Message    string
}

type MemoryRepository struct {
	mu sync.RWMutex

	persistencePath string

	nextUserID int64
	nextLogID  int64

	usersByID       map[int64]model.User
	usernames       map[string]int64
	tokens          map[string]int64
	playerStatuses  map[int64]model.PlayerStatus
	areas           map[int64]model.Area
	items           map[int64]model.Item
	dropsByArea     map[int64][]WeightedDrop
	inventories     map[int64]map[int64]int
	encyclopedia    map[int64]map[int64]time.Time
	notices         []model.Notice
	explorationLogs []model.ExplorationLog
	loginBonusClaim map[int64]time.Time

	cards                    map[int64]model.CharacterCard
	userCards                map[int64][]model.UserCharacterCard
	bosses                   map[int64]model.Boss
	checkpoints              map[string]model.Checkpoint
	checkpointByQR           map[string]string
	userCheckpointRecords    map[int64]map[string]model.UserCheckpointRecord
	userGachaTickets         map[int64]int
	userBossChallengeTickets map[int64]int
}

type persistentState struct {
	NextUserID               int64                                           `json:"next_user_id"`
	NextLogID                int64                                           `json:"next_log_id"`
	UsersByID                map[int64]model.User                            `json:"users_by_id"`
	PasswordHashes           map[int64]string                                `json:"password_hashes"`
	Tokens                   map[string]int64                                `json:"tokens"`
	PlayerStatuses           map[int64]model.PlayerStatus                    `json:"player_statuses"`
	Inventories              map[int64]map[int64]int                         `json:"inventories"`
	Encyclopedia             map[int64]map[int64]time.Time                   `json:"encyclopedia"`
	ExplorationLogs          []model.ExplorationLog                          `json:"exploration_logs"`
	LoginBonusClaim          map[int64]time.Time                             `json:"login_bonus_claim"`
	UserCards                map[int64][]model.UserCharacterCard             `json:"user_cards"`
	UserCheckpointRecords    map[int64]map[string]model.UserCheckpointRecord `json:"user_checkpoint_records"`
	UserGachaTickets         map[int64]int                                   `json:"user_gacha_tickets"`
	UserBossChallengeTickets map[int64]int                                   `json:"user_boss_challenge_tickets"`
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		nextUserID:               1,
		nextLogID:                1,
		usersByID:                make(map[int64]model.User),
		usernames:                make(map[string]int64),
		tokens:                   make(map[string]int64),
		playerStatuses:           make(map[int64]model.PlayerStatus),
		areas:                    make(map[int64]model.Area),
		items:                    make(map[int64]model.Item),
		dropsByArea:              make(map[int64][]WeightedDrop),
		inventories:              make(map[int64]map[int64]int),
		encyclopedia:             make(map[int64]map[int64]time.Time),
		loginBonusClaim:          make(map[int64]time.Time),
		cards:                    make(map[int64]model.CharacterCard),
		userCards:                make(map[int64][]model.UserCharacterCard),
		bosses:                   make(map[int64]model.Boss),
		checkpoints:              make(map[string]model.Checkpoint),
		checkpointByQR:           make(map[string]string),
		userCheckpointRecords:    make(map[int64]map[string]model.UserCheckpointRecord),
		userGachaTickets:         make(map[int64]int),
		userBossChallengeTickets: make(map[int64]int),
	}
}

func (r *MemoryRepository) SetPersistencePath(path string) {
	r.persistencePath = path
}

func (r *MemoryRepository) LoadPersistentState() error {
	if r.persistencePath == "" {
		return nil
	}

	data, err := os.ReadFile(r.persistencePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}

	var state persistentState
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if state.NextUserID > 0 {
		r.nextUserID = state.NextUserID
	}
	if state.NextLogID > 0 {
		r.nextLogID = state.NextLogID
	}
	if state.UsersByID != nil {
		r.usersByID = state.UsersByID
		r.usernames = make(map[string]int64, len(state.UsersByID))
		for id, user := range state.UsersByID {
			if state.PasswordHashes != nil {
				user.PasswordHash = state.PasswordHashes[id]
				r.usersByID[id] = user
			}
			r.usernames[user.Username] = id
		}
	}
	if state.Tokens != nil {
		r.tokens = state.Tokens
	}
	if state.PlayerStatuses != nil {
		r.playerStatuses = state.PlayerStatuses
	}
	if state.Inventories != nil {
		r.inventories = state.Inventories
	}
	if state.Encyclopedia != nil {
		r.encyclopedia = state.Encyclopedia
	}
	if state.ExplorationLogs != nil {
		r.explorationLogs = state.ExplorationLogs
	}
	if state.LoginBonusClaim != nil {
		r.loginBonusClaim = state.LoginBonusClaim
	}
	if state.UserCards != nil {
		r.userCards = state.UserCards
	}
	if state.UserCheckpointRecords != nil {
		r.userCheckpointRecords = state.UserCheckpointRecords
	}
	if state.UserGachaTickets != nil {
		r.userGachaTickets = state.UserGachaTickets
	}
	if state.UserBossChallengeTickets != nil {
		r.userBossChallengeTickets = state.UserBossChallengeTickets
	}

	return nil
}

func (r *MemoryRepository) savePersistentStateLocked() error {
	if r.persistencePath == "" {
		return nil
	}

	state := persistentState{
		NextUserID:               r.nextUserID,
		NextLogID:                r.nextLogID,
		UsersByID:                r.usersByID,
		PasswordHashes:           make(map[int64]string, len(r.usersByID)),
		Tokens:                   r.tokens,
		PlayerStatuses:           r.playerStatuses,
		Inventories:              r.inventories,
		Encyclopedia:             r.encyclopedia,
		ExplorationLogs:          r.explorationLogs,
		LoginBonusClaim:          r.loginBonusClaim,
		UserCards:                r.userCards,
		UserCheckpointRecords:    r.userCheckpointRecords,
		UserGachaTickets:         r.userGachaTickets,
		UserBossChallengeTickets: r.userBossChallengeTickets,
	}
	for id, user := range r.usersByID {
		state.PasswordHashes[id] = user.PasswordHash
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(r.persistencePath), 0o700); err != nil {
		return err
	}

	tmp := r.persistencePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, r.persistencePath)
}

func (r *MemoryRepository) CreateUser(username, passwordHash string) (model.User, model.PlayerStatus, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.usernames[username]; exists {
		return model.User{}, model.PlayerStatus{}, errors.New("username already exists")
	}
	now := time.Now()
	user := model.User{ID: r.nextUserID, Username: username, PasswordHash: passwordHash, CreatedAt: now, UpdatedAt: now}
	status := model.PlayerStatus{UserID: user.ID, Level: 1, Exp: 0, Stamina: -1, MaxStamina: -1, Coins: 500, Gems: 0, TotalExplorations: 0, LastStaminaRecoveredAt: now, UpdatedAt: now}
	r.usersByID[user.ID] = user
	r.usernames[username] = user.ID
	r.playerStatuses[user.ID] = status
	r.inventories[user.ID] = make(map[int64]int)
	r.encyclopedia[user.ID] = make(map[int64]time.Time)
	r.userCheckpointRecords[user.ID] = make(map[string]model.UserCheckpointRecord)

	starterIDs := make([]int64, 0)
	for _, c := range r.cards {
		if c.IsStarter {
			starterIDs = append(starterIDs, c.ID)
		}
	}
	sort.Slice(starterIDs, func(i, j int) bool { return starterIDs[i] < starterIDs[j] })
	for i, cardID := range starterIDs {
		r.userCards[user.ID] = append(r.userCards[user.ID], model.UserCharacterCard{UserID: user.ID, CardID: cardID, IsEquipped: i == 0, DeckSlot: i + 1, Level: 1, AcquiredAt: now})
	}
	r.nextUserID++
	_ = r.savePersistentStateLocked()
	return user, status, nil
}

func (r *MemoryRepository) Login(username, passwordHash string) (string, model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	uid, ok := r.usernames[username]
	if !ok {
		return "", model.User{}, ErrInvalidCredentials
	}
	user := r.usersByID[uid]
	if user.PasswordHash != passwordHash {
		return "", model.User{}, ErrInvalidCredentials
	}
	token, err := randomToken()
	if err != nil {
		return "", model.User{}, err
	}
	r.tokens[token] = user.ID
	_ = r.savePersistentStateLocked()
	return token, user, nil
}

func (r *MemoryRepository) GetUserByUsername(username string) (model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	uid, ok := r.usernames[username]
	if !ok {
		return model.User{}, ErrInvalidCredentials
	}
	return r.usersByID[uid], nil
}

func (r *MemoryRepository) IssueToken(userID int64) (string, model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, ok := r.usersByID[userID]
	if !ok {
		return "", model.User{}, ErrUserNotFound
	}
	token, err := randomToken()
	if err != nil {
		return "", model.User{}, err
	}
	r.tokens[token] = user.ID
	_ = r.savePersistentStateLocked()
	return token, user, nil
}

func (r *MemoryRepository) UpdatePasswordHash(userID int64, passwordHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, ok := r.usersByID[userID]
	if !ok {
		return ErrUserNotFound
	}
	user.PasswordHash = passwordHash
	user.UpdatedAt = time.Now()
	r.usersByID[userID] = user
	_ = r.savePersistentStateLocked()
	return nil
}

func (r *MemoryRepository) UserIDByToken(token string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	uid, ok := r.tokens[token]
	if !ok {
		return 0, ErrUnauthorized
	}
	return uid, nil
}
func (r *MemoryRepository) GetUserByID(userID int64) (model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.usersByID[userID]
	if !ok {
		return model.User{}, ErrUserNotFound
	}
	return u, nil
}
func (r *MemoryRepository) GetPlayerStatus(userID int64) (model.PlayerStatus, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	st, ok := r.playerStatuses[userID]
	if !ok {
		return model.PlayerStatus{}, ErrUserNotFound
	}
	return st, nil
}
func (r *MemoryRepository) SavePlayerStatus(status model.PlayerStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	status.UpdatedAt = time.Now()
	r.playerStatuses[status.UserID] = status
	_ = r.savePersistentStateLocked()
	return nil
}

func (r *MemoryRepository) ListAreas() []model.Area {
	r.mu.RLock()
	defer r.mu.RUnlock()
	areas := make([]model.Area, 0, len(r.areas))
	for _, a := range r.areas {
		if a.IsActive {
			areas = append(areas, a)
		}
	}
	sort.Slice(areas, func(i, j int) bool { return areas[i].SortOrder < areas[j].SortOrder })
	return areas
}
func (r *MemoryRepository) GetArea(areaID int64) (model.Area, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.areas[areaID]
	if !ok || !a.IsActive {
		return model.Area{}, ErrAreaNotFound
	}
	return a, nil
}
func (r *MemoryRepository) GetDrops(areaID int64) ([]WeightedDrop, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.dropsByArea[areaID]
	if !ok {
		return nil, ErrAreaNotFound
	}
	cp := make([]WeightedDrop, len(d))
	copy(cp, d)
	return cp, nil
}
func (r *MemoryRepository) GetItem(itemID int64) (model.Item, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	it, ok := r.items[itemID]
	if !ok {
		return model.Item{}, errors.New("item not found")
	}
	return it, nil
}
func (r *MemoryRepository) ListAllItems() []model.Item {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]model.Item, 0, len(r.items))
	for _, it := range r.items {
		items = append(items, it)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items
}
func (r *MemoryRepository) AddItemToInventory(userID, itemID int64, qty int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.inventories[userID]; !ok {
		r.inventories[userID] = make(map[int64]int)
	}
	r.inventories[userID][itemID] += qty
	_ = r.savePersistentStateLocked()
	return nil
}
func (r *MemoryRepository) ListInventory(userID int64) ([]model.InventoryEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	inv := r.inventories[userID]
	list := make([]model.InventoryEntry, 0, len(inv))
	for itemID, qty := range inv {
		if it, ok := r.items[itemID]; ok {
			list = append(list, model.InventoryEntry{ItemID: itemID, Name: it.Name, Rarity: it.Rarity, Quantity: qty})
		}
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].Rarity == list[j].Rarity {
			return list[i].ItemID < list[j].ItemID
		}
		return list[i].Rarity > list[j].Rarity
	})
	return list, nil
}
func (r *MemoryRepository) RegisterEncyclopedia(userID, itemID int64) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.encyclopedia[userID]; !ok {
		r.encyclopedia[userID] = make(map[int64]time.Time)
	}
	if _, exists := r.encyclopedia[userID][itemID]; exists {
		return false, nil
	}
	r.encyclopedia[userID][itemID] = time.Now()
	_ = r.savePersistentStateLocked()
	return true, nil
}
func (r *MemoryRepository) ListEncyclopedia(userID int64) ([]model.EncyclopediaEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entries := make([]model.EncyclopediaEntry, 0, len(r.items))
	userEntries := r.encyclopedia[userID]
	for _, it := range r.items {
		if !it.IsEncyclopediaTarget {
			continue
		}
		e := model.EncyclopediaEntry{ItemID: it.ID, Name: "？？？？", Rarity: it.Rarity, Obtained: false}
		if t, ok := userEntries[it.ID]; ok {
			e.Name = it.Name
			e.Obtained = true
			e.FirstObtainedAt = &t
		}
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].ItemID < entries[j].ItemID })
	return entries, nil
}
func (r *MemoryRepository) AddExplorationLog(log model.ExplorationLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	log.ID = r.nextLogID
	r.nextLogID++
	r.explorationLogs = append(r.explorationLogs, log)
	_ = r.savePersistentStateLocked()
	return nil
}
func (r *MemoryRepository) ClaimLoginBonus(userID int64) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	last, ok := r.loginBonusClaim[userID]
	if ok && sameDate(last, now) {
		return 0, ErrAlreadyClaimed
	}
	st := r.playerStatuses[userID]
	st.Coins += 100
	r.playerStatuses[userID] = st
	r.loginBonusClaim[userID] = now
	_ = r.savePersistentStateLocked()
	return 100, nil
}
func (r *MemoryRepository) ListNotices() []model.Notice {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]model.Notice, 0, len(r.notices))
	for _, n := range r.notices {
		if n.IsActive {
			out = append(out, n)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].IsPinned != out[j].IsPinned {
			return out[i].IsPinned
		}
		return out[i].PublishedAt.After(out[j].PublishedAt)
	})
	return out
}

func applyCardBonuses(card model.CharacterCard, uc model.UserCharacterCard) model.CharacterCard {
	card.MaxHP += uc.BonusHP
	card.Attack += uc.BonusAttack
	card.Defense += uc.BonusDefense
	return card
}
func (r *MemoryRepository) GetEquippedCard(userID int64) (model.CharacterCard, model.UserCharacterCard, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, uc := range r.userCards[userID] {
		if uc.IsEquipped {
			if c, ok := r.cards[uc.CardID]; ok {
				return applyCardBonuses(c, uc), uc, nil
			}
		}
	}
	return model.CharacterCard{}, model.UserCharacterCard{}, ErrCardNotFound
}
func (r *MemoryRepository) ListDeckCards(userID int64) ([]model.DeckCardView, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ucs := r.userCards[userID]
	if len(ucs) == 0 {
		return nil, ErrCardNotFound
	}
	out := make([]model.DeckCardView, 0, 6)
	for _, uc := range ucs {
		if uc.DeckSlot <= 0 {
			continue
		}
		if c, ok := r.cards[uc.CardID]; ok {
			out = append(out, model.DeckCardView{Card: applyCardBonuses(c, uc), DeckSlot: uc.DeckSlot, UpgradeLevel: uc.Level})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DeckSlot < out[j].DeckSlot })
	return out, nil
}
func (r *MemoryRepository) ListOwnedCards(userID int64) ([]model.CardCollectionEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]model.CardCollectionEntry, 0, len(r.userCards[userID]))
	for _, uc := range r.userCards[userID] {
		if c, ok := r.cards[uc.CardID]; ok {
			out = append(out, model.CardCollectionEntry{Card: applyCardBonuses(c, uc), User: uc, InDeck: uc.DeckSlot > 0})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].InDeck != out[j].InDeck {
			return out[i].InDeck
		}
		if out[i].User.DeckSlot != out[j].User.DeckSlot {
			return out[i].User.DeckSlot < out[j].User.DeckSlot
		}
		return out[i].Card.ID < out[j].Card.ID
	})
	return out, nil
}
func (r *MemoryRepository) ListCardArchive(userID int64) ([]model.CardArchiveEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	owned := make(map[int64]model.UserCharacterCard, len(r.userCards[userID]))
	for _, uc := range r.userCards[userID] {
		owned[uc.CardID] = uc
	}
	out := make([]model.CardArchiveEntry, 0, len(r.cards))
	for _, c := range r.cards {
		entry := model.CardArchiveEntry{Card: c}
		if uc, ok := owned[c.ID]; ok {
			entry.Card = applyCardBonuses(c, uc)
			entry.Obtained = true
			entry.InDeck = uc.DeckSlot > 0
			entry.DeckSlot = uc.DeckSlot
			entry.UpgradeLevel = uc.Level
		}
		out = append(out, entry)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Card.ID < out[j].Card.ID })
	return out, nil
}
func (r *MemoryRepository) UserHasCard(userID, cardID int64) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, uc := range r.userCards[userID] {
		if uc.CardID == cardID {
			return true
		}
	}
	return false
}
func (r *MemoryRepository) AddCardToUser(userID, cardID int64) (model.CharacterCard, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	card, ok := r.cards[cardID]
	if !ok {
		return model.CharacterCard{}, false, ErrCardNotFound
	}
	for i, uc := range r.userCards[userID] {
		if uc.CardID == cardID {
			uc.BonusHP += 1
			uc.BonusAttack += 1
			uc.BonusDefense += 1
			r.userCards[userID][i] = uc
			_ = r.savePersistentStateLocked()
			return applyCardBonuses(card, uc), true, nil
		}
	}
	used := map[int]bool{}
	for _, uc := range r.userCards[userID] {
		if uc.DeckSlot > 0 {
			used[uc.DeckSlot] = true
		}
	}
	deckSlot := 0
	for i := 1; i <= 6; i++ {
		if !used[i] {
			deckSlot = i
			break
		}
	}
	uc := model.UserCharacterCard{UserID: userID, CardID: cardID, DeckSlot: deckSlot, Level: 1, AcquiredAt: time.Now()}
	r.userCards[userID] = append(r.userCards[userID], uc)
	_ = r.savePersistentStateLocked()
	return applyCardBonuses(card, uc), false, nil
}
func (r *MemoryRepository) UpgradeCard(userID, cardID int64) (model.CharacterCard, model.UserCharacterCard, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	st, ok := r.playerStatuses[userID]
	if !ok {
		return model.CharacterCard{}, model.UserCharacterCard{}, 0, ErrUserNotFound
	}
	for i, uc := range r.userCards[userID] {
		if uc.CardID != cardID {
			continue
		}
		cost := 50 * uc.Level
		if st.Coins < cost {
			return model.CharacterCard{}, model.UserCharacterCard{}, cost, errors.New("not enough coins")
		}
		st.Coins -= cost
		uc.Level += 1
		uc.BonusHP += 2
		uc.BonusAttack += 1
		uc.BonusDefense += 1
		r.userCards[userID][i] = uc
		r.playerStatuses[userID] = st
		_ = r.savePersistentStateLocked()
		return applyCardBonuses(r.cards[cardID], uc), uc, cost, nil
	}
	return model.CharacterCard{}, model.UserCharacterCard{}, 0, ErrCardNotFound
}
func (r *MemoryRepository) UpdateDeck(userID int64, cardIDs []int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(cardIDs) != 6 {
		return errors.New("deck must contain 6 card ids")
	}
	seen := map[int64]bool{}
	for _, id := range cardIDs {
		if seen[id] {
			return errors.New("duplicate card in deck")
		}
		seen[id] = true
	}
	owned := map[int64]int{}
	for i, uc := range r.userCards[userID] {
		owned[uc.CardID] = i
		r.userCards[userID][i].DeckSlot = 0
		r.userCards[userID][i].IsEquipped = false
	}
	for slot, id := range cardIDs {
		idx, ok := owned[id]
		if !ok {
			return errors.New("deck includes unowned card")
		}
		r.userCards[userID][idx].DeckSlot = slot + 1
		if slot == 0 {
			r.userCards[userID][idx].IsEquipped = true
		}
	}
	_ = r.savePersistentStateLocked()
	return nil
}
func (r *MemoryRepository) ListGachaPool() []model.CharacterCard {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]model.CharacterCard, 0, len(r.cards))
	for _, c := range r.cards {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Rarity == out[j].Rarity {
			return out[i].ID < out[j].ID
		}
		return out[i].Rarity > out[j].Rarity
	})
	return out
}
func (r *MemoryRepository) SpendCoins(userID int64, amount int) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	st, ok := r.playerStatuses[userID]
	if !ok {
		return 0, ErrUserNotFound
	}
	if st.Coins < amount {
		return st.Coins, errors.New("not enough coins")
	}
	st.Coins -= amount
	r.playerStatuses[userID] = st
	_ = r.savePersistentStateLocked()
	return st.Coins, nil
}
func (r *MemoryRepository) AddCoins(userID int64, amount int) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	st, ok := r.playerStatuses[userID]
	if !ok {
		return 0, ErrUserNotFound
	}
	st.Coins += amount
	r.playerStatuses[userID] = st
	_ = r.savePersistentStateLocked()
	return st.Coins, nil
}
func (r *MemoryRepository) ListRewardCandidateCards() []model.CharacterCard {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]model.CharacterCard, 0)
	for _, c := range r.cards {
		if c.IsStarter {
			continue
		}
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
func (r *MemoryRepository) GetBoss(bossID int64) (model.Boss, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	boss, ok := r.bosses[bossID]
	if !ok {
		return model.Boss{}, ErrBossNotFound
	}
	return boss, nil
}

func (r *MemoryRepository) ListBosses() []model.Boss {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]model.Boss, 0, len(r.bosses))
	for _, boss := range r.bosses {
		out = append(out, boss)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (r *MemoryRepository) SeedAreas(areas []model.Area) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, a := range areas {
		r.areas[a.ID] = a
	}
}
func (r *MemoryRepository) SeedItems(items []model.Item) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, it := range items {
		r.items[it.ID] = it
	}
}
func (r *MemoryRepository) SeedDrops(areaID int64, drops []WeightedDrop) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.dropsByArea[areaID] = drops
}
func (r *MemoryRepository) SeedNotices(notices []model.Notice) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.notices = append(r.notices, notices...)
}
func (r *MemoryRepository) SeedCards(cards []model.CharacterCard) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range cards {
		r.cards[c.ID] = c
	}
}
func (r *MemoryRepository) SeedBosses(bosses []model.Boss) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, b := range bosses {
		r.bosses[b.ID] = b
	}
}

func (r *MemoryRepository) SeedCheckpoints(checkpoints []model.Checkpoint) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, cp := range checkpoints {
		r.checkpoints[cp.ID] = cp
		r.checkpointByQR[cp.QRText] = cp.ID
	}
}

func (r *MemoryRepository) ListCheckpoints() []model.Checkpoint {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]model.Checkpoint, 0, len(r.checkpoints))
	for _, cp := range r.checkpoints {
		if cp.IsActive {
			out = append(out, cp)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		left, leftErr := strconv.Atoi(strings.TrimPrefix(out[i].QRText, "QR"))
		right, rightErr := strconv.Atoi(strings.TrimPrefix(out[j].QRText, "QR"))
		if leftErr == nil && rightErr == nil {
			return left < right
		}
		return out[i].QRText < out[j].QRText
	})
	return out
}

func (r *MemoryRepository) GetCheckpointByQR(qrText string) (model.Checkpoint, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.checkpointByQR[qrText]
	if !ok {
		return model.Checkpoint{}, errors.New("checkpoint not found")
	}
	cp, ok := r.checkpoints[id]
	if !ok {
		return model.Checkpoint{}, errors.New("checkpoint not found")
	}
	return cp, nil
}

func (r *MemoryRepository) GetCheckpointRecord(userID int64, checkpointID string) model.UserCheckpointRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if rec, ok := r.userCheckpointRecords[userID][checkpointID]; ok {
		return rec
	}
	return model.UserCheckpointRecord{UserID: userID, CheckpointID: checkpointID}
}

func (r *MemoryRepository) SaveCheckpointRecord(rec model.UserCheckpointRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.userCheckpointRecords[rec.UserID]; !ok {
		r.userCheckpointRecords[rec.UserID] = make(map[string]model.UserCheckpointRecord)
	}
	r.userCheckpointRecords[rec.UserID][rec.CheckpointID] = rec
	_ = r.savePersistentStateLocked()
	return nil
}

func (r *MemoryRepository) AddExp(userID int64, exp int) (model.PlayerStatus, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	st, ok := r.playerStatuses[userID]
	if !ok {
		return model.PlayerStatus{}, ErrUserNotFound
	}
	st.Exp += exp
	for st.Exp >= expNeededForNextLevel(st.Level) {
		st.Exp -= expNeededForNextLevel(st.Level)
		st.Level++
	}
	r.playerStatuses[userID] = st
	_ = r.savePersistentStateLocked()
	return st, nil
}

func expNeededForNextLevel(level int) int {
	return 10 + (level-1)*10
}

func (r *MemoryRepository) AddGachaTickets(userID int64, amount int) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.userGachaTickets[userID] += amount
	_ = r.savePersistentStateLocked()
	return r.userGachaTickets[userID]
}
func (r *MemoryRepository) AddBossChallengeTickets(userID int64, amount int) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.userBossChallengeTickets[userID] += amount
	_ = r.savePersistentStateLocked()
	return r.userBossChallengeTickets[userID]
}
func (r *MemoryRepository) GetGachaTickets(userID int64) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.userGachaTickets[userID]
}
func (r *MemoryRepository) GetBossChallengeTickets(userID int64) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.userBossChallengeTickets[userID]
}

func NewDrop(itemID *int64, resultType string, weight, expReward, coinReward int, message string) WeightedDrop {
	return WeightedDrop{ItemID: itemID, ResultType: resultType, Weight: weight, ExpReward: expReward, CoinReward: coinReward, Message: message}
}
func randomToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
func sameDate(a, b time.Time) bool {
	y1, m1, d1 := a.Date()
	y2, m2, d2 := b.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
