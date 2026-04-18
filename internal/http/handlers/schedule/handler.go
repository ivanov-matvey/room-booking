package schedulehandler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ivanov-matvey/room-booking/internal/domain"
	"github.com/ivanov-matvey/room-booking/internal/http/response"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type UseCase interface {
	CreateSchedule(ctx context.Context, roomID uuid.UUID, daysOfWeek []int, startTime, endTime string) (*domain.Schedule, error)
}

type Handler struct {
	uc UseCase
}

func New(uc UseCase) *Handler {
	return &Handler{uc: uc}
}

// CreateSchedule godoc
// @Summary Создать расписание переговорки (только admin, только один раз). Длительность слота 30 мин. После создания расписание изменить нельзя.
// @Description Доступно только роли admin. Расписание можно создать только один раз для каждой переговорки. Поле `daysOfWeek` должно содержать значения от 1 до 7 (1=Пн, 7=Вс). При передаче значений вне этого диапазона возвращается 400.
// @Tags Schedules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param roomId path string true "Идентификатор переговорки"
// @Param request body ScheduleResponse true "Данные расписания"
// @Success 201 {object} CreateScheduleResponse "Расписание сохранено"
// @Failure 400 {object} response.ErrorResponse "Неверный запрос (в т.ч. недопустимые значения daysOfWeek)"
// @Failure 401 {object} response.ErrorResponse "Не авторизован"
// @Failure 403 {object} response.ErrorResponse "Доступ запрещён (требуется роль admin)"
// @Failure 404 {object} response.ErrorResponse "Переговорка не найдена"
// @Failure 409 {object} response.ErrorResponse "Расписание для переговорки уже создано, изменение не допускается"
// @Failure 500 {object} response.InternalErrorResponse "Внутренняя ошибка сервера"
// @Router /rooms/{roomId}/schedule/create [post]
func (h *Handler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	roomIDStr := chi.URLParam(r, "roomId")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid roomId")
		return
	}

	var req CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if len(req.DaysOfWeek) == 0 {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "daysOfWeek is required")
		return
	}
	for _, d := range req.DaysOfWeek {
		if d < 1 || d > 7 {
			response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "daysOfWeek values must be between 1 and 7")
			return
		}
	}
	if !domain.IsValidTime(req.StartTime) || !domain.IsValidTime(req.EndTime) {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "startTime and endTime must be in HH:MM format")
		return
	}
	if !domain.IsValidTimeRange(req.StartTime, req.EndTime) {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "startTime must be before endTime")
		return
	}

	schedule, err := h.uc.CreateSchedule(r.Context(), roomID, req.DaysOfWeek, req.StartTime, req.EndTime)
	if err != nil {
		response.DomainError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, CreateScheduleResponse{Schedule: new(domainScheduleToResponse(schedule))})
}

func domainScheduleToResponse(s *domain.Schedule) ScheduleResponse {
	return ScheduleResponse{
		ID:         new(s.ID),
		RoomID:     s.RoomID,
		DaysOfWeek: s.DaysOfWeek,
		StartTime:  s.StartTime,
		EndTime:    s.EndTime,
	}
}
