package handler

import (
	"net/http"

	"retro-treasure-backend/internal/service"
)

type AreaHandler struct{ svc *service.AreaService }

func NewAreaHandler(svc *service.AreaService) *AreaHandler {
	return &AreaHandler{svc: svc}
}

func (h *AreaHandler) List(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.svc.ListAreas())
}
