package slothandler

import (
	"context"
	"net/http"
	"time"

	"github.com/ivanov-matvey/room-booking/internal/domain"
	"github.com/ivanov-matvey/room-booking/internal/http/response"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type UseCase interface {
	GetAvailableSlots(ctx context.Context, roomID uuid.UUID, date time.Time) ([]domain.Slot, error)
}

type Handler struct {
	uc UseCase
}

func New(uc UseCase) *Handler {
	return &Handler{uc: uc}
}

// ListSlots godoc
// @Summary Список доступных для бронирования слотов по переговорке и дате (admin и user). Наиболее нагруженный эндпоинт.
// @Description Возвращает слоты, не занятые активной бронью, для указанной переговорки на указанную дату. Все даты и время передаются и возвращаются в UTC. Параметр `date` является обязательным; при его отсутствии возвращается 400. Если у переговорки нет расписания — возвращается пустой список (переговорка считается всегда недоступной).
// @Tags Slots
// @Produce json
// @Security BearerAuth
// @Param roomId path string true "Идентификатор переговорки"
// @Param date query string true "Дата в формате ISO 8601 (например: 2024-06-10). Обязательный параметр; при отсутствии — 400."
// @Success 200 {object} ListSlotsResponse "Список доступных слотов (не занятых активной бронью). Пустой список, если у переговорки нет расписания или на эту дату нет слотов."
// @Failure 400 {object} response.ErrorResponse "Неверный запрос (отсутствует или некорректен параметр date)"
// @Failure 401 {object} response.ErrorResponse "Не авторизован"
// @Failure 404 {object} response.ErrorResponse "Переговорка не найдена"
// @Failure 500 {object} response.InternalErrorResponse "Внутренняя ошибка сервера"
// @Router /rooms/{roomId}/slots/list [get]
func (h *Handler) ListSlots(w http.ResponseWriter, r *http.Request) {
	roomIDStr := chi.URLParam(r, "roomId")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid roomId")
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "date parameter is required")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid date format, expected YYYY-MM-DD")
		return
	}

	slots, err := h.uc.GetAvailableSlots(r.Context(), roomID, date)
	if err != nil {
		response.DomainError(w, err)
		return
	}

	result := make([]SlotResponse, 0, len(slots))
	for _, slot := range slots {
		result = append(result, domainSlotToResponse(slot))
	}

	response.JSON(w, http.StatusOK, ListSlotsResponse{Slots: result})
}

func domainSlotToResponse(s domain.Slot) SlotResponse {
	return SlotResponse{
		ID:     s.ID,
		RoomID: s.RoomID,
		Start:  s.StartTime.UTC(),
		End:    s.EndTime.UTC(),
	}
}
