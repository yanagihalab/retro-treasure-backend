package handler

import (
	"net/http"

	"retro-treasure-backend/internal/middleware"
	"retro-treasure-backend/internal/repository"
	"retro-treasure-backend/internal/service"
)

type LoginBonusHandler struct{ svc *service.LoginBonusService }

func NewLoginBonusHandler(svc *service.LoginBonusService) *LoginBonusHandler {
	return &LoginBonusHandler{svc: svc}
}

func (h *LoginBonusHandler) Claim(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	res, err := h.svc.Claim(userID)
	if err != nil {
		if err == repository.ErrAlreadyClaimed {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}
