package handler

import (
	"net/http"

	"retro-treasure-backend/internal/middleware"
	"retro-treasure-backend/internal/service"
)

type PlayerHandler struct{ svc *service.PlayerService }

func NewPlayerHandler(svc *service.PlayerService) *PlayerHandler { return &PlayerHandler{svc: svc} }

func (h *PlayerHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	res, err := h.svc.GetMe(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}
