package handler

import (
	"net/http"

	"retro-treasure-backend/internal/middleware"
	"retro-treasure-backend/internal/service"
)

type BossHandler struct{ svc *service.BossService }
func NewBossHandler(svc *service.BossService) *BossHandler { return &BossHandler{svc: svc} }

func (h *BossHandler) GetBoss(w http.ResponseWriter, r *http.Request) {
	res, err := h.svc.GetBoss()
	if err != nil { writeError(w, http.StatusInternalServerError, err.Error()); return }
	writeJSON(w, http.StatusOK, res)
}

func (h *BossHandler) AutoBattle(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok { writeError(w, http.StatusUnauthorized, "unauthorized"); return }
	res, err := h.svc.AutoBattle(userID)
	if err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
	writeJSON(w, http.StatusOK, res)
}
