package infohandler

import (
	"net/http"

	"github.com/ivanov-matvey/room-booking/internal/http/response"
)

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

// GetInfo godoc
// @Summary Статус сервиса
// @Description Возвращает статус сервиса.
// @Tags Info
// @Produce json
// @Success 200 {object} InfoResponse
// @Router /_info [get]
func (h *Handler) GetInfo(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
