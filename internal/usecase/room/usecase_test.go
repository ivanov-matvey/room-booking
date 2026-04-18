package room_test

import (
	"context"
	"testing"

	"github.com/ivanov-matvey/room-booking/internal/domain"
	roomusecase "github.com/ivanov-matvey/room-booking/internal/usecase/room"
	mockroom "github.com/ivanov-matvey/room-booking/mocks/usecase/room"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateRoom_Success(t *testing.T) {
	repo := mockroom.NewMockRepository(t)
	uc := roomusecase.New(repo)

	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Room")).Return(nil)

	desc := "A nice conference room"
	capacity := 10
	room, err := uc.CreateRoom(context.Background(), "Room A", &desc, &capacity)

	require.NoError(t, err)
	assert.NotNil(t, room)
	assert.Equal(t, "Room A", room.Name)
	assert.Equal(t, &desc, room.Description)
	assert.Equal(t, &capacity, room.Capacity)

	repo.AssertExpectations(t)
}

func TestCreateRoom_NoOptionalFields(t *testing.T) {
	repo := mockroom.NewMockRepository(t)
	uc := roomusecase.New(repo)

	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Room")).Return(nil)

	room, err := uc.CreateRoom(context.Background(), "Minimal Room", nil, nil)

	require.NoError(t, err)
	assert.NotNil(t, room)
	assert.Equal(t, "Minimal Room", room.Name)
	assert.Nil(t, room.Description)
	assert.Nil(t, room.Capacity)
}

func TestListRooms_Success(t *testing.T) {
	repo := mockroom.NewMockRepository(t)
	uc := roomusecase.New(repo)

	rooms := []domain.Room{
		{ID: uuid.New(), Name: "Room A"},
		{ID: uuid.New(), Name: "Room B"},
	}

	repo.On("List", mock.Anything).Return(rooms, nil)

	result, err := uc.ListRooms(context.Background())

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Room A", result[0].Name)
	assert.Equal(t, "Room B", result[1].Name)
}

func TestListRooms_Empty(t *testing.T) {
	repo := mockroom.NewMockRepository(t)
	uc := roomusecase.New(repo)

	repo.On("List", mock.Anything).Return([]domain.Room{}, nil)

	result, err := uc.ListRooms(context.Background())

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestListRooms_Error(t *testing.T) {
	repo := mockroom.NewMockRepository(t)
	uc := roomusecase.New(repo)

	repo.On("List", mock.Anything).Return(nil, assert.AnError)

	result, err := uc.ListRooms(context.Background())

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestCreateRoom_RepoError(t *testing.T) {
	repo := mockroom.NewMockRepository(t)
	uc := roomusecase.New(repo)

	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Room")).Return(assert.AnError)

	room, err := uc.CreateRoom(context.Background(), "Room A", nil, nil)

	require.Error(t, err)
	assert.Nil(t, room)
}
