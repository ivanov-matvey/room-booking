package roomhandler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ivanov-matvey/room-booking/internal/domain"
	roomhandler "github.com/ivanov-matvey/room-booking/internal/http/handlers/room"
	mockroomhandler "github.com/ivanov-matvey/room-booking/mocks/http/handlers/room"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListRooms_HappyPath(t *testing.T) {
	rooms := []domain.Room{
		{
			ID:          uuid.New(),
			Name:        "Room A",
			Description: new("board room"),
			Capacity:    new(10),
			CreatedAt:   time.Now(),
		},
	}

	m := mockroomhandler.NewMockUseCase(t)
	m.EXPECT().ListRooms(mock.Anything).Return(rooms, nil)

	h := roomhandler.New(m)

	req := httptest.NewRequest(http.MethodGet, "/rooms/list", nil)
	w := httptest.NewRecorder()
	h.ListRooms(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp roomhandler.ListRoomsResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Len(t, resp.Rooms, 1)
	assert.Equal(t, "Room A", resp.Rooms[0].Name)
}

func TestListRooms_Empty(t *testing.T) {
	m := mockroomhandler.NewMockUseCase(t)
	m.EXPECT().ListRooms(mock.Anything).Return([]domain.Room{}, nil)

	h := roomhandler.New(m)

	req := httptest.NewRequest(http.MethodGet, "/rooms/list", nil)
	w := httptest.NewRecorder()
	h.ListRooms(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp roomhandler.ListRoomsResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Empty(t, resp.Rooms)
}

func TestListRooms_InternalError(t *testing.T) {
	m := mockroomhandler.NewMockUseCase(t)
	m.EXPECT().ListRooms(mock.Anything).Return(nil, assert.AnError)

	h := roomhandler.New(m)

	req := httptest.NewRequest(http.MethodGet, "/rooms/list", nil)
	w := httptest.NewRecorder()
	h.ListRooms(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateRoom_HappyPath(t *testing.T) {
	room := &domain.Room{
		ID:          uuid.New(),
		Name:        "Room B",
		Description: new("a great room"),
		Capacity:    new(8),
		CreatedAt:   time.Now(),
	}

	m := mockroomhandler.NewMockUseCase(t)
	m.EXPECT().CreateRoom(mock.Anything, "Room B", mock.Anything, mock.Anything).Return(room, nil)

	h := roomhandler.New(m)

	body := bytes.NewBufferString(`{"name":"Room B","description":"a great room","capacity":8}`)
	req := httptest.NewRequest(http.MethodPost, "/rooms/create", body)
	w := httptest.NewRecorder()
	h.CreateRoom(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp roomhandler.CreateRoomResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.NotNil(t, resp.Room)
	assert.Equal(t, "Room B", resp.Room.Name)
}

func TestCreateRoom_MissingName(t *testing.T) {
	m := mockroomhandler.NewMockUseCase(t)

	h := roomhandler.New(m)

	body := bytes.NewBufferString(`{"name":""}`)
	req := httptest.NewRequest(http.MethodPost, "/rooms/create", body)
	w := httptest.NewRecorder()
	h.CreateRoom(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateRoom_InvalidJSON(t *testing.T) {
	m := mockroomhandler.NewMockUseCase(t)

	h := roomhandler.New(m)

	body := bytes.NewBufferString(`not-json`)
	req := httptest.NewRequest(http.MethodPost, "/rooms/create", body)
	w := httptest.NewRecorder()
	h.CreateRoom(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateRoom_InternalError(t *testing.T) {
	m := mockroomhandler.NewMockUseCase(t)
	m.EXPECT().CreateRoom(mock.Anything, "Room C", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	h := roomhandler.New(m)

	body := bytes.NewBufferString(`{"name":"Room C"}`)
	req := httptest.NewRequest(http.MethodPost, "/rooms/create", body)
	w := httptest.NewRecorder()
	h.CreateRoom(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
