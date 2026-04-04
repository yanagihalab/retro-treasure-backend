package handler

import (
	"net/http"

	"retro-treasure-backend/internal/middleware"
	"retro-treasure-backend/internal/service"
)

type ItemHandler struct{ svc *service.ItemService }

func NewItemHandler(svc *service.ItemService) *ItemHandler { return &ItemHandler{svc: svc} }

func (h *ItemHandler) Inventory(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	res, err := h.svc.ListInventory(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *ItemHandler) Encyclopedia(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	res, err := h.svc.GetEncyclopedia(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}
