package handler

import (
	"net/http"

	"retro-treasure-backend/internal/service"
)

type NoticeHandler struct{ svc *service.NoticeService }

func NewNoticeHandler(svc *service.NoticeService) *NoticeHandler {
	return &NoticeHandler{svc: svc}
}

func (h *NoticeHandler) List(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.svc.ListNotices())
}
