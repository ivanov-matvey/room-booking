package authhandler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ivanov-matvey/room-booking/internal/domain"
	authhandler "github.com/ivanov-matvey/room-booking/internal/http/handlers/auth"
	mockauthhandler "github.com/ivanov-matvey/room-booking/mocks/http/handlers/auth"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDummyLogin_HappyPath(t *testing.T) {
	m := mockauthhandler.NewMockUseCase(t)
	m.EXPECT().DummyLogin(mock.Anything, "user").Return("test-token", nil)

	h := authhandler.New(m)

	body := bytes.NewBufferString(`{"role":"user"}`)
	req := httptest.NewRequest(http.MethodPost, "/dummyLogin", body)
	w := httptest.NewRecorder()
	h.DummyLogin(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp authhandler.TokenResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "test-token", resp.Token)
}

func TestDummyLogin_AdminRole(t *testing.T) {
	m := mockauthhandler.NewMockUseCase(t)
	m.EXPECT().DummyLogin(mock.Anything, "admin").Return("admin-token", nil)

	h := authhandler.New(m)

	body := bytes.NewBufferString(`{"role":"admin"}`)
	req := httptest.NewRequest(http.MethodPost, "/dummyLogin", body)
	w := httptest.NewRecorder()
	h.DummyLogin(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDummyLogin_InvalidRole(t *testing.T) {
	m := mockauthhandler.NewMockUseCase(t)

	h := authhandler.New(m)

	body := bytes.NewBufferString(`{"role":"superuser"}`)
	req := httptest.NewRequest(http.MethodPost, "/dummyLogin", body)
	w := httptest.NewRecorder()
	h.DummyLogin(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDummyLogin_InvalidJSON(t *testing.T) {
	m := mockauthhandler.NewMockUseCase(t)

	h := authhandler.New(m)

	body := bytes.NewBufferString(`not-json`)
	req := httptest.NewRequest(http.MethodPost, "/dummyLogin", body)
	w := httptest.NewRecorder()
	h.DummyLogin(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDummyLogin_UCError(t *testing.T) {
	m := mockauthhandler.NewMockUseCase(t)
	m.EXPECT().DummyLogin(mock.Anything, "user").Return("", domain.ErrInvalidRequest)

	h := authhandler.New(m)

	body := bytes.NewBufferString(`{"role":"user"}`)
	req := httptest.NewRequest(http.MethodPost, "/dummyLogin", body)
	w := httptest.NewRecorder()
	h.DummyLogin(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_HappyPath(t *testing.T) {
	now := time.Now()
	user := &domain.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Role:      "user",
		CreatedAt: now,
	}

	m := mockauthhandler.NewMockUseCase(t)
	m.EXPECT().Register(mock.Anything, "test@example.com", "password123", "user").Return(user, nil)

	h := authhandler.New(m)

	body := bytes.NewBufferString(`{"email":"test@example.com","password":"password123","role":"user"}`)
	req := httptest.NewRequest(http.MethodPost, "/register", body)
	w := httptest.NewRecorder()
	h.Register(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp authhandler.RegisterResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.NotNil(t, resp.User)
	assert.Equal(t, "test@example.com", resp.User.Email)
}

func TestRegister_MissingEmail(t *testing.T) {
	m := mockauthhandler.NewMockUseCase(t)

	h := authhandler.New(m)

	body := bytes.NewBufferString(`{"email":"","password":"password123","role":"user"}`)
	req := httptest.NewRequest(http.MethodPost, "/register", body)
	w := httptest.NewRecorder()
	h.Register(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_MissingPassword(t *testing.T) {
	m := mockauthhandler.NewMockUseCase(t)

	h := authhandler.New(m)

	body := bytes.NewBufferString(`{"email":"test@example.com","password":"","role":"user"}`)
	req := httptest.NewRequest(http.MethodPost, "/register", body)
	w := httptest.NewRecorder()
	h.Register(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_InvalidEmail(t *testing.T) {
	m := mockauthhandler.NewMockUseCase(t)

	h := authhandler.New(m)

	body := bytes.NewBufferString(`{"email":"notanemail","password":"password123","role":"user"}`)
	req := httptest.NewRequest(http.MethodPost, "/register", body)
	w := httptest.NewRecorder()
	h.Register(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_ShortPassword(t *testing.T) {
	m := mockauthhandler.NewMockUseCase(t)

	h := authhandler.New(m)

	body := bytes.NewBufferString(`{"email":"test@example.com","password":"abc","role":"user"}`)
	req := httptest.NewRequest(http.MethodPost, "/register", body)
	w := httptest.NewRecorder()
	h.Register(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_InvalidRole(t *testing.T) {
	m := mockauthhandler.NewMockUseCase(t)

	h := authhandler.New(m)

	body := bytes.NewBufferString(`{"email":"test@example.com","password":"password123","role":"superuser"}`)
	req := httptest.NewRequest(http.MethodPost, "/register", body)
	w := httptest.NewRecorder()
	h.Register(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_InvalidJSON(t *testing.T) {
	m := mockauthhandler.NewMockUseCase(t)

	h := authhandler.New(m)

	body := bytes.NewBufferString(`not-json`)
	req := httptest.NewRequest(http.MethodPost, "/register", body)
	w := httptest.NewRecorder()
	h.Register(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_UCError_InternalError(t *testing.T) {
	m := mockauthhandler.NewMockUseCase(t)
	m.EXPECT().Register(mock.Anything, "test@example.com", "password123", "user").Return(nil, assert.AnError)

	h := authhandler.New(m)

	body := bytes.NewBufferString(`{"email":"test@example.com","password":"password123","role":"user"}`)
	req := httptest.NewRequest(http.MethodPost, "/register", body)
	w := httptest.NewRecorder()
	h.Register(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestLogin_HappyPath(t *testing.T) {
	m := mockauthhandler.NewMockUseCase(t)
	m.EXPECT().Login(mock.Anything, "test@example.com", "password123").Return("login-token", nil)

	h := authhandler.New(m)

	body := bytes.NewBufferString(`{"email":"test@example.com","password":"password123"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	w := httptest.NewRecorder()
	h.Login(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp authhandler.TokenResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "login-token", resp.Token)
}

func TestLogin_InvalidJSON(t *testing.T) {
	m := mockauthhandler.NewMockUseCase(t)

	h := authhandler.New(m)

	body := bytes.NewBufferString(`not-json`)
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	w := httptest.NewRecorder()
	h.Login(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_Unauthorized(t *testing.T) {
	m := mockauthhandler.NewMockUseCase(t)
	m.EXPECT().Login(mock.Anything, "test@example.com", "wrongpass").Return("", domain.ErrUnauthorized)

	h := authhandler.New(m)

	body := bytes.NewBufferString(`{"email":"test@example.com","password":"wrongpass"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	w := httptest.NewRecorder()
	h.Login(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogin_InternalError(t *testing.T) {
	m := mockauthhandler.NewMockUseCase(t)
	m.EXPECT().Login(mock.Anything, "test@example.com", "password123").Return("", assert.AnError)

	h := authhandler.New(m)

	body := bytes.NewBufferString(`{"email":"test@example.com","password":"password123"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	w := httptest.NewRecorder()
	h.Login(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
