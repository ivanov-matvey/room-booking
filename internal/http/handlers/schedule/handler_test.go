package schedulehandler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ivanov-matvey/room-booking/internal/domain"
	schedulehandler "github.com/ivanov-matvey/room-booking/internal/http/handlers/schedule"
	mockschedulehandler "github.com/ivanov-matvey/room-booking/mocks/http/handlers/schedule"

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

func TestCreateSchedule_HappyPath(t *testing.T) {
	roomID := uuid.New()
	schedID := uuid.New()
	schedule := &domain.Schedule{
		ID:         schedID,
		RoomID:     roomID,
		DaysOfWeek: []int{1, 2, 3},
		StartTime:  "09:00",
		EndTime:    "18:00",
	}

	m := mockschedulehandler.NewMockUseCase(t)
	m.EXPECT().CreateSchedule(mock.Anything, roomID, []int{1, 2, 3}, "09:00", "18:00").Return(schedule, nil)

	h := schedulehandler.New(m)

	body := bytes.NewBufferString(`{"daysOfWeek":[1,2,3],"startTime":"09:00","endTime":"18:00"}`)
	req := httptest.NewRequest(http.MethodPost, "/rooms/"+roomID.String()+"/schedule/create", body)
	req = withChiParams(req, map[string]string{"roomId": roomID.String()})
	w := httptest.NewRecorder()
	h.CreateSchedule(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp schedulehandler.CreateScheduleResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.NotNil(t, resp.Schedule)
	assert.Equal(t, "09:00", resp.Schedule.StartTime)
	assert.Equal(t, "18:00", resp.Schedule.EndTime)
}

func TestCreateSchedule_InvalidRoomID(t *testing.T) {
	m := mockschedulehandler.NewMockUseCase(t)

	h := schedulehandler.New(m)

	body := bytes.NewBufferString(`{"daysOfWeek":[1,2,3],"startTime":"09:00","endTime":"18:00"}`)
	req := httptest.NewRequest(http.MethodPost, "/rooms/not-a-uuid/schedule/create", body)
	req = withChiParams(req, map[string]string{"roomId": "not-a-uuid"})
	w := httptest.NewRecorder()
	h.CreateSchedule(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateSchedule_InvalidJSON(t *testing.T) {
	roomID := uuid.New()
	m := mockschedulehandler.NewMockUseCase(t)

	h := schedulehandler.New(m)

	body := bytes.NewBufferString(`not-json`)
	req := httptest.NewRequest(http.MethodPost, "/rooms/"+roomID.String()+"/schedule/create", body)
	req = withChiParams(req, map[string]string{"roomId": roomID.String()})
	w := httptest.NewRecorder()
	h.CreateSchedule(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateSchedule_EmptyDaysOfWeek(t *testing.T) {
	roomID := uuid.New()
	m := mockschedulehandler.NewMockUseCase(t)

	h := schedulehandler.New(m)

	body := bytes.NewBufferString(`{"daysOfWeek":[],"startTime":"09:00","endTime":"18:00"}`)
	req := httptest.NewRequest(http.MethodPost, "/rooms/"+roomID.String()+"/schedule/create", body)
	req = withChiParams(req, map[string]string{"roomId": roomID.String()})
	w := httptest.NewRecorder()
	h.CreateSchedule(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateSchedule_InvalidDayOfWeek(t *testing.T) {
	roomID := uuid.New()
	m := mockschedulehandler.NewMockUseCase(t)

	h := schedulehandler.New(m)

	body := bytes.NewBufferString(`{"daysOfWeek":[0,8],"startTime":"09:00","endTime":"18:00"}`)
	req := httptest.NewRequest(http.MethodPost, "/rooms/"+roomID.String()+"/schedule/create", body)
	req = withChiParams(req, map[string]string{"roomId": roomID.String()})
	w := httptest.NewRecorder()
	h.CreateSchedule(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateSchedule_InvalidTimeFormat(t *testing.T) {
	roomID := uuid.New()
	m := mockschedulehandler.NewMockUseCase(t)

	h := schedulehandler.New(m)

	body := bytes.NewBufferString(`{"daysOfWeek":[1,2],"startTime":"9am","endTime":"18:00"}`)
	req := httptest.NewRequest(http.MethodPost, "/rooms/"+roomID.String()+"/schedule/create", body)
	req = withChiParams(req, map[string]string{"roomId": roomID.String()})
	w := httptest.NewRecorder()
	h.CreateSchedule(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateSchedule_StartAfterEnd(t *testing.T) {
	roomID := uuid.New()
	m := mockschedulehandler.NewMockUseCase(t)

	h := schedulehandler.New(m)

	body := bytes.NewBufferString(`{"daysOfWeek":[1,2],"startTime":"18:00","endTime":"09:00"}`)
	req := httptest.NewRequest(http.MethodPost, "/rooms/"+roomID.String()+"/schedule/create", body)
	req = withChiParams(req, map[string]string{"roomId": roomID.String()})
	w := httptest.NewRecorder()
	h.CreateSchedule(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateSchedule_ScheduleExists(t *testing.T) {
	roomID := uuid.New()
	m := mockschedulehandler.NewMockUseCase(t)
	m.EXPECT().CreateSchedule(mock.Anything, roomID, []int{1, 2, 3}, "09:00", "18:00").Return(nil, domain.ErrScheduleExists)

	h := schedulehandler.New(m)

	body := bytes.NewBufferString(`{"daysOfWeek":[1,2,3],"startTime":"09:00","endTime":"18:00"}`)
	req := httptest.NewRequest(http.MethodPost, "/rooms/"+roomID.String()+"/schedule/create", body)
	req = withChiParams(req, map[string]string{"roomId": roomID.String()})
	w := httptest.NewRecorder()
	h.CreateSchedule(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestCreateSchedule_RoomNotFound(t *testing.T) {
	roomID := uuid.New()
	m := mockschedulehandler.NewMockUseCase(t)
	m.EXPECT().CreateSchedule(mock.Anything, roomID, []int{1, 2, 3}, "09:00", "18:00").Return(nil, domain.ErrRoomNotFound)

	h := schedulehandler.New(m)

	body := bytes.NewBufferString(`{"daysOfWeek":[1,2,3],"startTime":"09:00","endTime":"18:00"}`)
	req := httptest.NewRequest(http.MethodPost, "/rooms/"+roomID.String()+"/schedule/create", body)
	req = withChiParams(req, map[string]string{"roomId": roomID.String()})
	w := httptest.NewRecorder()
	h.CreateSchedule(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateSchedule_InternalError(t *testing.T) {
	roomID := uuid.New()
	m := mockschedulehandler.NewMockUseCase(t)
	m.EXPECT().CreateSchedule(mock.Anything, roomID, []int{1, 2, 3}, "09:00", "18:00").Return(nil, assert.AnError)

	h := schedulehandler.New(m)

	body := bytes.NewBufferString(`{"daysOfWeek":[1,2,3],"startTime":"09:00","endTime":"18:00"}`)
	req := httptest.NewRequest(http.MethodPost, "/rooms/"+roomID.String()+"/schedule/create", body)
	req = withChiParams(req, map[string]string{"roomId": roomID.String()})
	w := httptest.NewRecorder()
	h.CreateSchedule(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
