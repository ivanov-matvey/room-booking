package authhandler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ivanov-matvey/room-booking/internal/domain"
	"github.com/ivanov-matvey/room-booking/internal/http/response"
)

type UseCase interface {
	DummyLogin(ctx context.Context, role string) (string, error)
	Register(ctx context.Context, email, password, role string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (string, error)
}

type Handler struct {
	uc UseCase
}

func New(uc UseCase) *Handler {
	return &Handler{uc: uc}
}

// DummyLogin godoc
// @Summary Получить тестовый JWT по роли. Доступен без авторизации.
// @Description Выдаёт тестовый JWT для указанной роли (admin / user). Для каждой роли возвращается **фиксированный UUID** пользователя: один и тот же UUID для всех запросов с ролью admin и один и тот же UUID для роли user. Это обеспечивает стабильность при тестировании сценариев, требующих проверки владельца брони. JWT содержит `user_id` (UUID) и `role`. Принимает только допустимые значения роли: `admin` или `user`; иные значения — 400.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body DummyLoginRequest true "Роль пользователя"
// @Success 200 {object} TokenResponse "Тестовый токен"
// @Failure 400 {object} response.ErrorResponse "Неверный запрос (недопустимое значение роли)"
// @Failure 500 {object} response.InternalErrorResponse "Внутренняя ошибка сервера"
// @Router /dummyLogin [post]
func (h *Handler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	var req DummyLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.Role != domain.RoleAdmin && req.Role != domain.RoleUser {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "role must be 'admin' or 'user'")
		return
	}

	token, err := h.uc.DummyLogin(r.Context(), req.Role)
	if err != nil {
		response.DomainError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, TokenResponse{Token: token})
}

// Register godoc
// @Summary Регистрация пользователя
// @Description Создаёт нового пользователя и возвращает его данные. Доступен без авторизации.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Данные нового пользователя"
// @Success 201 {object} RegisterResponse "Пользователь создан"
// @Failure 400 {object} response.ErrorResponse "Неверный запрос или email уже занят"
// @Failure 500 {object} response.InternalErrorResponse "Внутренняя ошибка сервера"
// @Router /register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "email and password are required")
		return
	}
	if !strings.Contains(req.Email, "@") || strings.HasPrefix(req.Email, "@") || strings.HasSuffix(req.Email, "@") {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid email format")
		return
	}
	if len(req.Password) < 6 {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "password must be at least 6 characters")
		return
	}
	if req.Role != domain.RoleAdmin && req.Role != domain.RoleUser {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "role must be 'admin' or 'user'")
		return
	}

	user, err := h.uc.Register(r.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		response.DomainError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, RegisterResponse{User: domainUserToResponse(user)})
}

// Login godoc
// @Summary Авторизация по email и паролю
// @Description Авторизует пользователя по email и паролю, возвращает JWT. Доступен без авторизации.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Учётные данные"
// @Success 200 {object} TokenResponse "Успешная авторизация"
// @Failure 401 {object} response.ErrorResponse "Неверные учётные данные"
// @Failure 500 {object} response.InternalErrorResponse "Внутренняя ошибка сервера"
// @Router /login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	token, err := h.uc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		response.DomainError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, TokenResponse{Token: token})
}

func domainUserToResponse(u *domain.User) *UserResponse {
	if u == nil {
		return nil
	}
	return &UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: new(u.CreatedAt),
	}
}
