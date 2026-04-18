package slot_test

import (
	"context"
	"testing"
	"time"

	"github.com/ivanov-matvey/room-booking/internal/domain"
	slotusecase "github.com/ivanov-matvey/room-booking/internal/usecase/slot"
	mockslot "github.com/ivanov-matvey/room-booking/mocks/usecase/slot"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetAvailableSlots_Success(t *testing.T) {
	slotRepo := mockslot.NewMockRepository(t)
	roomRepo := mockslot.NewMockRoomRepository(t)
	scheduleRepo := mockslot.NewMockScheduleRepository(t)

	uc := slotusecase.New(slotRepo, roomRepo, scheduleRepo)

	roomID := uuid.New()
	date := time.Now().AddDate(0, 0, 1).UTC()
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, 1)
	}

	room := &domain.Room{ID: roomID, Name: "Test Room"}
	schedule := &domain.Schedule{
		ID:         uuid.New(),
		RoomID:     roomID,
		DaysOfWeek: []int{1, 2, 3, 4, 5},
		StartTime:  "09:00",
		EndTime:    "18:00",
	}

	from := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 0, 1)

	slots := []domain.Slot{
		{ID: uuid.New(), RoomID: roomID, StartTime: from.Add(9 * time.Hour), EndTime: from.Add(9*time.Hour + 30*time.Minute)},
	}

	roomRepo.On("GetByID", mock.Anything, roomID).Return(room, nil)
	scheduleRepo.On("GetByRoomID", mock.Anything, roomID).Return(schedule, nil)
	slotRepo.On("CountByRoomAndDate", mock.Anything, roomID, from, to).Return(1, nil)
	slotRepo.On("ListAvailable", mock.Anything, roomID, from, to).Return(slots, nil)

	result, err := uc.GetAvailableSlots(context.Background(), roomID, date)

	require.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestGetAvailableSlots_RoomNotFound(t *testing.T) {
	slotRepo := mockslot.NewMockRepository(t)
	roomRepo := mockslot.NewMockRoomRepository(t)
	scheduleRepo := mockslot.NewMockScheduleRepository(t)

	uc := slotusecase.New(slotRepo, roomRepo, scheduleRepo)

	roomID := uuid.New()
	date := time.Now().AddDate(0, 0, 1)

	roomRepo.On("GetByID", mock.Anything, roomID).Return(nil, domain.ErrRoomNotFound)

	result, err := uc.GetAvailableSlots(context.Background(), roomID, date)

	require.Error(t, err)
	assert.Equal(t, domain.ErrRoomNotFound, err)
	assert.Nil(t, result)
}

func TestGetAvailableSlots_NoSchedule(t *testing.T) {
	slotRepo := mockslot.NewMockRepository(t)
	roomRepo := mockslot.NewMockRoomRepository(t)
	scheduleRepo := mockslot.NewMockScheduleRepository(t)

	uc := slotusecase.New(slotRepo, roomRepo, scheduleRepo)

	roomID := uuid.New()
	date := time.Now().AddDate(0, 0, 1)

	room := &domain.Room{ID: roomID, Name: "Test Room"}

	roomRepo.On("GetByID", mock.Anything, roomID).Return(room, nil)
	scheduleRepo.On("GetByRoomID", mock.Anything, roomID).Return(nil, nil)

	result, err := uc.GetAvailableSlots(context.Background(), roomID, date)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestGetAvailableSlots_GenerateSlots(t *testing.T) {
	slotRepo := mockslot.NewMockRepository(t)
	roomRepo := mockslot.NewMockRoomRepository(t)
	scheduleRepo := mockslot.NewMockScheduleRepository(t)

	uc := slotusecase.New(slotRepo, roomRepo, scheduleRepo)

	roomID := uuid.New()
	date := time.Now().AddDate(0, 0, 1).UTC()
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, 1)
	}

	room := &domain.Room{ID: roomID, Name: "Test Room"}
	schedule := &domain.Schedule{
		ID:         uuid.New(),
		RoomID:     roomID,
		DaysOfWeek: []int{1},
		StartTime:  "09:00",
		EndTime:    "10:00",
	}

	from := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 0, 1)

	generatedSlots := []domain.Slot{
		{ID: uuid.New(), RoomID: roomID, StartTime: from.Add(9 * time.Hour), EndTime: from.Add(9*time.Hour + 30*time.Minute)},
		{ID: uuid.New(), RoomID: roomID, StartTime: from.Add(9*time.Hour + 30*time.Minute), EndTime: from.Add(10 * time.Hour)},
	}

	roomRepo.On("GetByID", mock.Anything, roomID).Return(room, nil)
	scheduleRepo.On("GetByRoomID", mock.Anything, roomID).Return(schedule, nil)
	slotRepo.On("CountByRoomAndDate", mock.Anything, roomID, from, to).Return(0, nil)
	slotRepo.On("BulkUpsert", mock.Anything, mock.AnythingOfType("[]domain.Slot")).Return(nil)
	slotRepo.On("ListAvailable", mock.Anything, roomID, from, to).Return(generatedSlots, nil)

	result, err := uc.GetAvailableSlots(context.Background(), roomID, date)

	require.NoError(t, err)
	assert.Len(t, result, 2)

	slotRepo.AssertCalled(t, "BulkUpsert", mock.Anything, mock.Anything)
}

func TestGetAvailableSlots_ScheduleRepoError(t *testing.T) {
	slotRepo := mockslot.NewMockRepository(t)
	roomRepo := mockslot.NewMockRoomRepository(t)
	scheduleRepo := mockslot.NewMockScheduleRepository(t)

	uc := slotusecase.New(slotRepo, roomRepo, scheduleRepo)

	roomID := uuid.New()
	date := time.Now().AddDate(0, 0, 1)
	room := &domain.Room{ID: roomID, Name: "Test Room"}

	roomRepo.On("GetByID", mock.Anything, roomID).Return(room, nil)
	scheduleRepo.On("GetByRoomID", mock.Anything, roomID).Return(nil, assert.AnError)

	result, err := uc.GetAvailableSlots(context.Background(), roomID, date)

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestGetAvailableSlots_CountError(t *testing.T) {
	slotRepo := mockslot.NewMockRepository(t)
	roomRepo := mockslot.NewMockRoomRepository(t)
	scheduleRepo := mockslot.NewMockScheduleRepository(t)

	uc := slotusecase.New(slotRepo, roomRepo, scheduleRepo)

	roomID := uuid.New()
	date := time.Now().AddDate(0, 0, 1).UTC()
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, 1)
	}

	room := &domain.Room{ID: roomID, Name: "Test Room"}
	schedule := &domain.Schedule{
		ID:         uuid.New(),
		RoomID:     roomID,
		DaysOfWeek: []int{1},
		StartTime:  "09:00",
		EndTime:    "10:00",
	}

	from := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 0, 1)

	roomRepo.On("GetByID", mock.Anything, roomID).Return(room, nil)
	scheduleRepo.On("GetByRoomID", mock.Anything, roomID).Return(schedule, nil)
	slotRepo.On("CountByRoomAndDate", mock.Anything, roomID, from, to).Return(0, assert.AnError)

	result, err := uc.GetAvailableSlots(context.Background(), roomID, date)

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestGetAvailableSlots_BulkUpsertError(t *testing.T) {
	slotRepo := mockslot.NewMockRepository(t)
	roomRepo := mockslot.NewMockRoomRepository(t)
	scheduleRepo := mockslot.NewMockScheduleRepository(t)

	uc := slotusecase.New(slotRepo, roomRepo, scheduleRepo)

	roomID := uuid.New()
	date := time.Now().AddDate(0, 0, 1).UTC()
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, 1)
	}

	room := &domain.Room{ID: roomID, Name: "Test Room"}
	schedule := &domain.Schedule{
		ID:         uuid.New(),
		RoomID:     roomID,
		DaysOfWeek: []int{1},
		StartTime:  "09:00",
		EndTime:    "10:00",
	}

	from := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 0, 1)

	roomRepo.On("GetByID", mock.Anything, roomID).Return(room, nil)
	scheduleRepo.On("GetByRoomID", mock.Anything, roomID).Return(schedule, nil)
	slotRepo.On("CountByRoomAndDate", mock.Anything, roomID, from, to).Return(0, nil)
	slotRepo.On("BulkUpsert", mock.Anything, mock.AnythingOfType("[]domain.Slot")).Return(assert.AnError)

	result, err := uc.GetAvailableSlots(context.Background(), roomID, date)

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestGetAvailableSlots_ListAvailableError(t *testing.T) {
	slotRepo := mockslot.NewMockRepository(t)
	roomRepo := mockslot.NewMockRoomRepository(t)
	scheduleRepo := mockslot.NewMockScheduleRepository(t)

	uc := slotusecase.New(slotRepo, roomRepo, scheduleRepo)

	roomID := uuid.New()
	date := time.Now().AddDate(0, 0, 1).UTC()
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, 1)
	}

	room := &domain.Room{ID: roomID, Name: "Test Room"}
	schedule := &domain.Schedule{
		ID:         uuid.New(),
		RoomID:     roomID,
		DaysOfWeek: []int{1},
		StartTime:  "09:00",
		EndTime:    "10:00",
	}

	from := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 0, 1)

	roomRepo.On("GetByID", mock.Anything, roomID).Return(room, nil)
	scheduleRepo.On("GetByRoomID", mock.Anything, roomID).Return(schedule, nil)
	slotRepo.On("CountByRoomAndDate", mock.Anything, roomID, from, to).Return(1, nil)
	slotRepo.On("ListAvailable", mock.Anything, roomID, from, to).Return(nil, assert.AnError)

	result, err := uc.GetAvailableSlots(context.Background(), roomID, date)

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestGetAvailableSlots_DateTooFarInFuture(t *testing.T) {
	slotRepo := mockslot.NewMockRepository(t)
	roomRepo := mockslot.NewMockRoomRepository(t)
	scheduleRepo := mockslot.NewMockScheduleRepository(t)

	uc := slotusecase.New(slotRepo, roomRepo, scheduleRepo)

	roomID := uuid.New()
	date := time.Now().AddDate(0, 0, 95).UTC()

	room := &domain.Room{ID: roomID, Name: "Test Room"}
	schedule := &domain.Schedule{
		ID:         uuid.New(),
		RoomID:     roomID,
		DaysOfWeek: []int{1, 2, 3, 4, 5, 6, 7},
		StartTime:  "09:00",
		EndTime:    "10:00",
	}

	from := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 0, 1)

	roomRepo.On("GetByID", mock.Anything, roomID).Return(room, nil)
	scheduleRepo.On("GetByRoomID", mock.Anything, roomID).Return(schedule, nil)
	slotRepo.On("ListAvailable", mock.Anything, roomID, from, to).Return([]domain.Slot{}, nil)

	result, err := uc.GetAvailableSlots(context.Background(), roomID, date)

	require.NoError(t, err)
	assert.Empty(t, result)
	slotRepo.AssertNotCalled(t, "CountByRoomAndDate")
}
