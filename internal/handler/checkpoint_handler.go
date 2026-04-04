package handler

import (
	"encoding/json"
	"net/http"

	"retro-treasure-backend/internal/middleware"
	"retro-treasure-backend/internal/model"
	"retro-treasure-backend/internal/service"
)

type CheckpointHandler struct{ svc *service.CheckpointService }

func NewCheckpointHandler(svc *service.CheckpointService) *CheckpointHandler {
	return &CheckpointHandler{svc: svc}
}

func (h *CheckpointHandler) Master(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.svc.GetMaster())
}
func (h *CheckpointHandler) History(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	writeJSON(w, http.StatusOK, h.svc.GetHistory(userID))
}
func (h *CheckpointHandler) Claim(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req model.CheckpointClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	res, err := h.svc.Claim(userID, req.QRText)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}
