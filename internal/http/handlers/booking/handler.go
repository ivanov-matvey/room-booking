package bookinghandler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ivanov-matvey/room-booking/internal/domain"
	"github.com/ivanov-matvey/room-booking/internal/http/middleware"
	"github.com/ivanov-matvey/room-booking/internal/http/response"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type UseCase interface {
	CreateBooking(ctx context.Context, userID uuid.UUID, slotID uuid.UUID, createConferenceLink bool) (*domain.Booking, error)
	CancelBooking(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) (*domain.Booking, error)
	GetMyBookings(ctx context.Context, userID uuid.UUID) ([]domain.Booking, error)
	ListBookings(ctx context.Context, page, pageSize int) ([]domain.Booking, int, error)
}

type Handler struct {
	uc UseCase
}

func New(uc UseCase) *Handler {
	return &Handler{uc: uc}
}

// CreateBooking godoc
// @Summary Создать бронь на слот (только user). Опционально — запросить ссылку на конференцию.
// @Description Доступно только роли user. Администратор не может создавать брони (403). `userId` берётся из JWT-токена, а не из тела запроса. Если временной слот находится в прошлом (start < now), возвращается 400 (INVALID_REQUEST).
// @Tags Bookings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateBookingRequest true "Данные брони"
// @Success 201 {object} CreateBookingResponse "Бронь создана"
// @Failure 400 {object} response.ErrorResponse "Неверный запрос"
// @Failure 401 {object} response.ErrorResponse "Не авторизован"
// @Failure 403 {object} response.ErrorResponse "Доступ запрещён (бронирование доступно только роли user)"
// @Failure 404 {object} response.ErrorResponse "Слот не найден"
// @Failure 409 {object} response.ErrorResponse "Слот уже занят"
// @Failure 500 {object} response.InternalErrorResponse "Внутренняя ошибка сервера"
// @Router /bookings/create [post]
func (h *Handler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.DomainError(w, domain.ErrUnauthorized)
		return
	}

	var req CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.SlotID == uuid.Nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "slotId is required")
		return
	}

	createConferenceLink := false
	if req.CreateConferenceLink != nil {
		createConferenceLink = *req.CreateConferenceLink
	}

	booking, err := h.uc.CreateBooking(r.Context(), userID, req.SlotID, createConferenceLink)
	if err != nil {
		response.DomainError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, CreateBookingResponse{Booking: new(domainBookingToResponse(*booking))})
}

// ListBookings godoc
// @Summary Список всех броней с пагинацией (только admin)
// @Description Доступно только роли admin. Поддерживает пагинацию через параметры `page` и `pageSize`. Оба параметра опциональны; значения по умолчанию: `page=1`, `pageSize=20`. Максимальное значение `pageSize` — 100.
// @Tags Bookings
// @Produce json
// @Security BearerAuth
// @Param page query int false "Номер страницы (начиная с 1). По умолчанию 1."
// @Param pageSize query int false "Количество записей на странице. По умолчанию 20, максимум 100."
// @Success 200 {object} ListBookingsResponse "Список всех броней"
// @Failure 400 {object} response.ErrorResponse "Неверный запрос (некорректные параметры пагинации)"
// @Failure 401 {object} response.ErrorResponse "Не авторизован"
// @Failure 403 {object} response.ErrorResponse "Доступ запрещён (только admin)"
// @Failure 500 {object} response.InternalErrorResponse "Внутренняя ошибка сервера"
// @Router /bookings/list [get]
func (h *Handler) ListBookings(w http.ResponseWriter, r *http.Request) {
	page := 1
	pageSize := 20

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil || p < 1 {
			response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid page parameter")
			return
		}
		page = p
	}
	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		ps, err := strconv.Atoi(pageSizeStr)
		if err != nil || ps < 1 || ps > 100 {
			response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid pageSize parameter (must be between 1 and 100)")
			return
		}
		pageSize = ps
	}

	bookings, total, err := h.uc.ListBookings(r.Context(), page, pageSize)
	if err != nil {
		response.DomainError(w, err)
		return
	}

	result := make([]BookingResponse, 0, len(bookings))
	for _, b := range bookings {
		result = append(result, domainBookingToResponse(b))
	}

	response.JSON(w, http.StatusOK, ListBookingsResponse{
		Bookings: result,
		Pagination: &PaginationResponse{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

// GetMyBookings godoc
// @Summary Список броней текущего пользователя (только user)
// @Description Доступно только роли user. Возвращает брони пользователя, чей `user_id` содержится в JWT-токене. Возвращаются только брони на будущие слоты (start >= now); брони на уже прошедшие слоты в ответ не включаются.
// @Tags Bookings
// @Produce json
// @Security BearerAuth
// @Success 200 {object} GetMyBookingsResponse "Список броней текущего пользователя"
// @Failure 401 {object} response.ErrorResponse "Не авторизован"
// @Failure 403 {object} response.ErrorResponse "Доступ запрещён (только user)"
// @Failure 500 {object} response.InternalErrorResponse "Внутренняя ошибка сервера"
// @Router /bookings/my [get]
func (h *Handler) GetMyBookings(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.DomainError(w, domain.ErrUnauthorized)
		return
	}

	bookings, err := h.uc.GetMyBookings(r.Context(), userID)
	if err != nil {
		response.DomainError(w, err)
		return
	}

	result := make([]BookingResponse, 0, len(bookings))
	for _, b := range bookings {
		result = append(result, domainBookingToResponse(b))
	}

	response.JSON(w, http.StatusOK, GetMyBookingsResponse{Bookings: result})
}

// CancelBooking godoc
// @Summary Отменить бронь (только своя бронь, только user)
// @Description Доступно только роли user. Пользователь может отменить только свою бронь. Операция **идемпотентна**: повторный вызов на уже отменённой брони не является ошибкой и возвращает 200 с актуальным состоянием брони (status: cancelled). При попытке отменить чужую бронь — 403. При отсутствии брони с указанным ID — 404.
// @Tags Bookings
// @Produce json
// @Security BearerAuth
// @Param bookingId path string true "Идентификатор брони"
// @Success 200 {object} CancelBookingResponse "Бронь отменена (или уже была отменена ранее)"
// @Failure 401 {object} response.ErrorResponse "Не авторизован"
// @Failure 403 {object} response.ErrorResponse "Не своя бронь или роль не user"
// @Failure 404 {object} response.ErrorResponse "Бронь не найдена"
// @Failure 500 {object} response.InternalErrorResponse "Внутренняя ошибка сервера"
// @Router /bookings/{bookingId}/cancel [post]
func (h *Handler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.DomainError(w, domain.ErrUnauthorized)
		return
	}

	bookingIDStr := chi.URLParam(r, "bookingId")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid bookingId")
		return
	}

	booking, err := h.uc.CancelBooking(r.Context(), userID, bookingID)
	if err != nil {
		response.DomainError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, CancelBookingResponse{Booking: new(domainBookingToResponse(*booking))})
}

func domainBookingToResponse(b domain.Booking) BookingResponse {
	return BookingResponse{
		ID:             b.ID,
		SlotID:         b.SlotID,
		UserID:         b.UserID,
		Status:         b.Status,
		ConferenceLink: b.ConferenceLink,
		CreatedAt:      new(b.CreatedAt),
	}
}
