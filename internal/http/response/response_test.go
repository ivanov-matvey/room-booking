package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ivanov-matvey/room-booking/internal/domain"
	"github.com/ivanov-matvey/room-booking/internal/http/response"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSON(t *testing.T) {
	w := httptest.NewRecorder()
	response.JSON(w, http.StatusOK, map[string]string{"key": "value"})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var result map[string]string
	require.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	assert.Equal(t, "value", result["key"])
}

func TestJSON_Created(t *testing.T) {
	w := httptest.NewRecorder()
	response.JSON(w, http.StatusCreated, map[string]int{"count": 42})

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "bad input")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var result map[string]map[string]string
	require.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	assert.Equal(t, "INVALID_REQUEST", result["error"]["code"])
	assert.Equal(t, "bad input", result["error"]["message"])
}

func TestDomainError_RoomNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	response.DomainError(w, domain.ErrRoomNotFound)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDomainError_SlotNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	response.DomainError(w, domain.ErrSlotNotFound)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDomainError_BookingNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	response.DomainError(w, domain.ErrBookingNotFound)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDomainError_SlotBooked(t *testing.T) {
	w := httptest.NewRecorder()
	response.DomainError(w, domain.ErrSlotBooked)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestDomainError_ScheduleExists(t *testing.T) {
	w := httptest.NewRecorder()
	response.DomainError(w, domain.ErrScheduleExists)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestDomainError_Forbidden(t *testing.T) {
	w := httptest.NewRecorder()
	response.DomainError(w, domain.ErrForbidden)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDomainError_Unauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	response.DomainError(w, domain.ErrUnauthorized)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDomainError_PastSlot(t *testing.T) {
	w := httptest.NewRecorder()
	response.DomainError(w, domain.ErrPastSlot)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDomainError_InvalidRequest(t *testing.T) {
	w := httptest.NewRecorder()
	response.DomainError(w, domain.ErrInvalidRequest)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDomainError_Internal(t *testing.T) {
	w := httptest.NewRecorder()
	response.DomainError(w, assert.AnError)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var result map[string]map[string]string
	require.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	assert.Equal(t, "INTERNAL_ERROR", result["error"]["code"])
}
