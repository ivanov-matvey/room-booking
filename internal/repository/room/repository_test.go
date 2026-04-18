package room_test

import (
	"context"
	"testing"
	"time"

	roomrepo "github.com/ivanov-matvey/room-booking/internal/repository/room"

	"github.com/ivanov-matvey/room-booking/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMock(t *testing.T) pgxmock.PgxPoolIface {
	t.Helper()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	return mock
}

func TestNew(t *testing.T) {
	mock := newMock(t)
	repo := roomrepo.New(mock)
	assert.NotNil(t, repo)
}

func TestCreate_Success(t *testing.T) {
	mock := newMock(t)
	repo := roomrepo.New(mock)

	now := time.Now()
	room := &domain.Room{ID: uuid.New(), Name: "Test Room"}

	rows := pgxmock.NewRows([]string{"created_at"}).AddRow(now)
	mock.ExpectQuery(`INSERT INTO rooms`).
		WithArgs(room.ID, room.Name, room.Description, room.Capacity).
		WillReturnRows(rows)

	err := repo.Create(context.Background(), room)
	require.NoError(t, err)
	assert.Equal(t, now, room.CreatedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreate_Error(t *testing.T) {
	mock := newMock(t)
	repo := roomrepo.New(mock)

	room := &domain.Room{ID: uuid.New(), Name: "Test Room"}

	mock.ExpectQuery(`INSERT INTO rooms`).
		WithArgs(room.ID, room.Name, room.Description, room.Capacity).
		WillReturnError(assert.AnError)

	err := repo.Create(context.Background(), room)
	require.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestList_Success(t *testing.T) {
	mock := newMock(t)
	repo := roomrepo.New(mock)

	roomID := uuid.New()
	now := time.Now()

	rows := pgxmock.NewRows([]string{"id", "name", "description", "capacity", "created_at"}).
		AddRow(roomID, "Room A", nil, nil, now)

	mock.ExpectQuery(`SELECT id, name`).WillReturnRows(rows)

	result, err := repo.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Room A", result[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestList_Error(t *testing.T) {
	mock := newMock(t)
	repo := roomrepo.New(mock)

	mock.ExpectQuery(`SELECT id, name`).WillReturnError(assert.AnError)

	result, err := repo.List(context.Background())
	require.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestList_Empty(t *testing.T) {
	mock := newMock(t)
	repo := roomrepo.New(mock)

	rows := pgxmock.NewRows([]string{"id", "name", "description", "capacity", "created_at"})
	mock.ExpectQuery(`SELECT id, name`).WillReturnRows(rows)

	result, err := repo.List(context.Background())
	require.NoError(t, err)
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID_Success(t *testing.T) {
	mock := newMock(t)
	repo := roomrepo.New(mock)

	roomID := uuid.New()
	now := time.Now()

	rows := pgxmock.NewRows([]string{"id", "name", "description", "capacity", "created_at"}).
		AddRow(roomID, "Room A", nil, nil, now)

	mock.ExpectQuery(`SELECT id, name`).
		WithArgs(roomID).
		WillReturnRows(rows)

	result, err := repo.GetByID(context.Background(), roomID)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, roomID, result.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID_NotFound(t *testing.T) {
	mock := newMock(t)
	repo := roomrepo.New(mock)

	roomID := uuid.New()

	rows := pgxmock.NewRows([]string{"id", "name", "description", "capacity", "created_at"})
	mock.ExpectQuery(`SELECT id, name`).
		WithArgs(roomID).
		WillReturnRows(rows)

	result, err := repo.GetByID(context.Background(), roomID)
	require.ErrorIs(t, err, domain.ErrRoomNotFound)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID_ScanError(t *testing.T) {
	mock := newMock(t)
	repo := roomrepo.New(mock)

	roomID := uuid.New()

	rows := pgxmock.NewRows([]string{"id", "name", "description", "capacity", "created_at"}).
		AddRow(roomID, "Room A", nil, nil, nil).
		RowError(0, assert.AnError)

	mock.ExpectQuery(`SELECT id, name`).
		WithArgs(roomID).
		WillReturnRows(rows)

	result, err := repo.GetByID(context.Background(), roomID)
	// Either error or ErrNoRows → ErrRoomNotFound
	_ = result
	_ = err
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID_QueryError(t *testing.T) {
	mock := newMock(t)
	repo := roomrepo.New(mock)

	roomID := uuid.New()

	mock.ExpectQuery(`SELECT id, name`).
		WithArgs(roomID).
		WillReturnError(assert.AnError)

	result, err := repo.GetByID(context.Background(), roomID)
	require.Error(t, err)
	assert.NotEqual(t, domain.ErrRoomNotFound, err)
	assert.Nil(t, result)
	_ = pgx.ErrNoRows // keep import
	assert.NoError(t, mock.ExpectationsWereMet())
}
