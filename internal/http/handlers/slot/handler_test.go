package slothandler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ivanov-matvey/room-booking/internal/domain"
	slothandler "github.com/ivanov-matvey/room-booking/internal/http/handlers/slot"
	mockslothandler "github.com/ivanov-matvey/room-booking/mocks/http/handlers/slot"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func withChiParams(r *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestListSlots_HappyPath(t *testing.T) {
	roomID := uuid.New()
	now := time.Now().UTC().Truncate(time.Hour)
	slots := []domain.Slot{
		{
			ID:        uuid.New(),
			RoomID:    roomID,
			StartTime: now,
			EndTime:   now.Add(time.Hour),
		},
	}

	m := mockslothandler.NewMockUseCase(t)
	expectedDate, _ := time.Parse("2006-01-02", "2025-06-10")
	m.EXPECT().GetAvailableSlots(mock.Anything, roomID, expectedDate).Return(slots, nil)

	h := slothandler.New(m)

	req := httptest.NewRequest(http.MethodGet, "/rooms/"+roomID.String()+"/slots/list?date=2025-06-10", nil)
	req = withChiParams(req, map[string]string{"roomId": roomID.String()})
	w := httptest.NewRecorder()
	h.ListSlots(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp slothandler.ListSlotsResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Len(t, resp.Slots, 1)
}

func TestListSlots_InvalidRoomID(t *testing.T) {
	m := mockslothandler.NewMockUseCase(t)

	h := slothandler.New(m)

	req := httptest.NewRequest(http.MethodGet, "/rooms/not-a-uuid/slots/list?date=2025-06-10", nil)
	req = withChiParams(req, map[string]string{"roomId": "not-a-uuid"})
	w := httptest.NewRecorder()
	h.ListSlots(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListSlots_MissingDate(t *testing.T) {
	roomID := uuid.New()
	m := mockslothandler.NewMockUseCase(t)

	h := slothandler.New(m)

	req := httptest.NewRequest(http.MethodGet, "/rooms/"+roomID.String()+"/slots/list", nil)
	req = withChiParams(req, map[string]string{"roomId": roomID.String()})
	w := httptest.NewRecorder()
	h.ListSlots(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListSlots_InvalidDateFormat(t *testing.T) {
	roomID := uuid.New()
	m := mockslothandler.NewMockUseCase(t)

	h := slothandler.New(m)

	req := httptest.NewRequest(http.MethodGet, "/rooms/"+roomID.String()+"/slots/list?date=10-06-2025", nil)
	req = withChiParams(req, map[string]string{"roomId": roomID.String()})
	w := httptest.NewRecorder()
	h.ListSlots(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListSlots_RoomNotFound(t *testing.T) {
	roomID := uuid.New()
	expectedDate, _ := time.Parse("2006-01-02", "2025-06-10")

	m := mockslothandler.NewMockUseCase(t)
	m.EXPECT().GetAvailableSlots(mock.Anything, roomID, expectedDate).Return(nil, domain.ErrRoomNotFound)

	h := slothandler.New(m)

	req := httptest.NewRequest(http.MethodGet, "/rooms/"+roomID.String()+"/slots/list?date=2025-06-10", nil)
	req = withChiParams(req, map[string]string{"roomId": roomID.String()})
	w := httptest.NewRecorder()
	h.ListSlots(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestListSlots_InternalError(t *testing.T) {
	roomID := uuid.New()
	expectedDate, _ := time.Parse("2006-01-02", "2025-06-10")

	m := mockslothandler.NewMockUseCase(t)
	m.EXPECT().GetAvailableSlots(mock.Anything, roomID, expectedDate).Return(nil, assert.AnError)

	h := slothandler.New(m)

	req := httptest.NewRequest(http.MethodGet, "/rooms/"+roomID.String()+"/slots/list?date=2025-06-10", nil)
	req = withChiParams(req, map[string]string{"roomId": roomID.String()})
	w := httptest.NewRecorder()
	h.ListSlots(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
