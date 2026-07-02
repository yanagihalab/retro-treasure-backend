package handler

import (
	"encoding/json"
	"net/http"

	"retro-treasure-backend/internal/middleware"
	"retro-treasure-backend/internal/model"
	"retro-treasure-backend/internal/repository"
	"retro-treasure-backend/internal/service"
)

type ExploreHandler struct{ svc *service.ExploreService }

func NewExploreHandler(svc *service.ExploreService) *ExploreHandler {
	return &ExploreHandler{svc: svc}
}

func (h *ExploreHandler) Explore(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req model.ExploreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	res, err := h.svc.Explore(userID, req)
	if err != nil {
		switch err {
		case repository.ErrAreaLocked, repository.ErrInsufficientStamina, repository.ErrAreaNotFound:
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, res)
}
