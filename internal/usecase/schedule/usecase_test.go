package schedule_test

import (
	"context"
	"testing"

	"github.com/ivanov-matvey/room-booking/internal/domain"
	scheduleusecase "github.com/ivanov-matvey/room-booking/internal/usecase/schedule"
	mockschedule "github.com/ivanov-matvey/room-booking/mocks/usecase/schedule"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateSchedule_Success(t *testing.T) {
	scheduleRepo := mockschedule.NewMockRepository(t)
	roomRepo := mockschedule.NewMockRoomRepository(t)

	uc := scheduleusecase.New(scheduleRepo, roomRepo)

	roomID := uuid.New()
	room := &domain.Room{ID: roomID, Name: "Test Room"}

	roomRepo.On("GetByID", mock.Anything, roomID).Return(room, nil)
	scheduleRepo.On("GetByRoomID", mock.Anything, roomID).Return(nil, nil)
	scheduleRepo.On("CreateWithSlots", mock.Anything, mock.AnythingOfType("*domain.Schedule"), mock.AnythingOfType("[]domain.Slot")).Return(nil)

	schedule, err := uc.CreateSchedule(context.Background(), roomID, []int{1, 2, 3, 4, 5}, "09:00", "17:00")

	require.NoError(t, err)
	assert.NotNil(t, schedule)
	assert.Equal(t, roomID, schedule.RoomID)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, schedule.DaysOfWeek)
	assert.Equal(t, "09:00", schedule.StartTime)
	assert.Equal(t, "17:00", schedule.EndTime)

	scheduleRepo.AssertExpectations(t)
	roomRepo.AssertExpectations(t)
}

func TestCreateSchedule_RoomNotFound(t *testing.T) {
	scheduleRepo := mockschedule.NewMockRepository(t)
	roomRepo := mockschedule.NewMockRoomRepository(t)

	uc := scheduleusecase.New(scheduleRepo, roomRepo)

	roomID := uuid.New()

	roomRepo.On("GetByID", mock.Anything, roomID).Return(nil, domain.ErrRoomNotFound)

	schedule, err := uc.CreateSchedule(context.Background(), roomID, []int{1}, "09:00", "17:00")

	require.Error(t, err)
	assert.Equal(t, domain.ErrRoomNotFound, err)
	assert.Nil(t, schedule)
}

func TestCreateSchedule_AlreadyExists(t *testing.T) {
	scheduleRepo := mockschedule.NewMockRepository(t)
	roomRepo := mockschedule.NewMockRoomRepository(t)

	uc := scheduleusecase.New(scheduleRepo, roomRepo)

	roomID := uuid.New()
	room := &domain.Room{ID: roomID, Name: "Test Room"}
	existing := &domain.Schedule{
		ID:         uuid.New(),
		RoomID:     roomID,
		DaysOfWeek: []int{1},
		StartTime:  "09:00",
		EndTime:    "17:00",
	}

	roomRepo.On("GetByID", mock.Anything, roomID).Return(room, nil)
	scheduleRepo.On("GetByRoomID", mock.Anything, roomID).Return(existing, nil)

	schedule, err := uc.CreateSchedule(context.Background(), roomID, []int{1}, "10:00", "18:00")

	require.Error(t, err)
	assert.Equal(t, domain.ErrScheduleExists, err)
	assert.Nil(t, schedule)
	scheduleRepo.AssertNotCalled(t, "CreateWithSlots")
}

func TestCreateSchedule_GetByRoomIDError(t *testing.T) {
	scheduleRepo := mockschedule.NewMockRepository(t)
	roomRepo := mockschedule.NewMockRoomRepository(t)

	uc := scheduleusecase.New(scheduleRepo, roomRepo)

	roomID := uuid.New()
	room := &domain.Room{ID: roomID, Name: "Test Room"}

	roomRepo.On("GetByID", mock.Anything, roomID).Return(room, nil)
	scheduleRepo.On("GetByRoomID", mock.Anything, roomID).Return(nil, assert.AnError)

	schedule, err := uc.CreateSchedule(context.Background(), roomID, []int{1}, "09:00", "17:00")

	require.Error(t, err)
	assert.Nil(t, schedule)
}

func TestCreateSchedule_CreateWithSlotsError(t *testing.T) {
	scheduleRepo := mockschedule.NewMockRepository(t)
	roomRepo := mockschedule.NewMockRoomRepository(t)

	uc := scheduleusecase.New(scheduleRepo, roomRepo)

	roomID := uuid.New()
	room := &domain.Room{ID: roomID, Name: "Test Room"}

	roomRepo.On("GetByID", mock.Anything, roomID).Return(room, nil)
	scheduleRepo.On("GetByRoomID", mock.Anything, roomID).Return(nil, nil)
	scheduleRepo.On("CreateWithSlots", mock.Anything, mock.AnythingOfType("*domain.Schedule"), mock.AnythingOfType("[]domain.Slot")).Return(assert.AnError)

	schedule, err := uc.CreateSchedule(context.Background(), roomID, []int{1}, "09:00", "17:00")

	require.Error(t, err)
	assert.Nil(t, schedule)
}

func TestGetScheduleByRoom_Success(t *testing.T) {
	scheduleRepo := mockschedule.NewMockRepository(t)
	roomRepo := mockschedule.NewMockRoomRepository(t)

	uc := scheduleusecase.New(scheduleRepo, roomRepo)

	roomID := uuid.New()
	expected := &domain.Schedule{ID: uuid.New(), RoomID: roomID, DaysOfWeek: []int{1}, StartTime: "09:00", EndTime: "17:00"}

	scheduleRepo.On("GetByRoomID", mock.Anything, roomID).Return(expected, nil)

	result, err := uc.GetScheduleByRoom(context.Background(), roomID)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestGetScheduleByRoom_Error(t *testing.T) {
	scheduleRepo := mockschedule.NewMockRepository(t)
	roomRepo := mockschedule.NewMockRoomRepository(t)

	uc := scheduleusecase.New(scheduleRepo, roomRepo)

	roomID := uuid.New()
	scheduleRepo.On("GetByRoomID", mock.Anything, roomID).Return(nil, domain.ErrRoomNotFound)

	result, err := uc.GetScheduleByRoom(context.Background(), roomID)

	require.Error(t, err)
	assert.Nil(t, result)
}
