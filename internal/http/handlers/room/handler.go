package roomhandler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ivanov-matvey/room-booking/internal/domain"
	"github.com/ivanov-matvey/room-booking/internal/http/response"
)

type UseCase interface {
	CreateRoom(ctx context.Context, name string, description *string, capacity *int) (*domain.Room, error)
	ListRooms(ctx context.Context) ([]domain.Room, error)
}

type Handler struct {
	uc UseCase
}

func New(uc UseCase) *Handler {
	return &Handler{uc: uc}
}

// ListRooms godoc
// @Summary Список переговорок (admin и user)
// @Tags Rooms
// @Produce json
// @Security BearerAuth
// @Success 200 {object} ListRoomsResponse "Список переговорок"
// @Failure 401 {object} response.ErrorResponse "Не авторизован"
// @Failure 500 {object} response.InternalErrorResponse "Внутренняя ошибка сервера"
// @Router /rooms/list [get]
func (h *Handler) ListRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.uc.ListRooms(r.Context())
	if err != nil {
		response.DomainError(w, err)
		return
	}

	result := make([]RoomResponse, 0, len(rooms))
	for _, room := range rooms {
		result = append(result, domainRoomToResponse(room))
	}

	response.JSON(w, http.StatusOK, ListRoomsResponse{Rooms: result})
}

// CreateRoom godoc
// @Summary Создать переговорку (только admin)
// @Tags Rooms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateRoomRequest true "Данные переговорки"
// @Success 201 {object} CreateRoomResponse "Переговорка создана"
// @Failure 400 {object} response.ErrorResponse "Неверный запрос"
// @Failure 401 {object} response.ErrorResponse "Не авторизован"
// @Failure 403 {object} response.ErrorResponse "Доступ запрещён (требуется роль admin)"
// @Failure 500 {object} response.InternalErrorResponse "Внутренняя ошибка сервера"
// @Router /rooms/create [post]
func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.Name == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "name is required")
		return
	}

	room, err := h.uc.CreateRoom(r.Context(), req.Name, req.Description, req.Capacity)
	if err != nil {
		response.DomainError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, CreateRoomResponse{Room: new(domainRoomToResponse(*room))})
}

func domainRoomToResponse(r domain.Room) RoomResponse {
	return RoomResponse{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		Capacity:    r.Capacity,
		CreatedAt:   new(r.CreatedAt),
	}
}
