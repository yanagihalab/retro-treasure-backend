package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"retro-treasure-backend/internal/middleware"
	"retro-treasure-backend/internal/service"
)

type BossHandler struct{ svc *service.BossService }

func NewBossHandler(svc *service.BossService) *BossHandler { return &BossHandler{svc: svc} }

func (h *BossHandler) GetBoss(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	bossID := parseBossID(r)
	res, err := h.svc.GetBoss(userID, bossID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *BossHandler) AutoBattle(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	bossID := parseBossID(r)
	if r.Body != nil {
		var req struct {
			BossID int64 `json:"boss_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.BossID > 0 {
			bossID = req.BossID
		}
	}
	res, err := h.svc.AutoBattle(userID, bossID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func parseBossID(r *http.Request) int64 {
	raw := r.URL.Query().Get("id")
	if raw == "" {
		raw = r.URL.Query().Get("boss_id")
	}
	if raw == "" {
		return 1
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 1
	}
	return id
}
