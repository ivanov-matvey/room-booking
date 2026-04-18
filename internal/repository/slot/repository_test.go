package slot_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ivanov-matvey/room-booking/internal/domain"
	slotrepo "github.com/ivanov-matvey/room-booking/internal/repository/slot"
)

func newMock(t *testing.T) pgxmock.PgxPoolIface {
	t.Helper()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	return mock
}

func TestNew(t *testing.T) {
	mock := newMock(t)
	repo := slotrepo.New(mock)
	assert.NotNil(t, repo)
}

func TestGetByID_Success(t *testing.T) {
	mock := newMock(t)
	repo := slotrepo.New(mock)

	slotID := uuid.New()
	roomID := uuid.New()
	now := time.Now()

	rows := pgxmock.NewRows([]string{"id", "room_id", "start_time", "end_time", "created_at"}).
		AddRow(slotID, roomID, now, now.Add(30*time.Minute), now)

	mock.ExpectQuery(`SELECT id, room_id`).
		WithArgs(slotID).
		WillReturnRows(rows)

	result, err := repo.GetByID(context.Background(), slotID)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, slotID, result.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID_NotFound(t *testing.T) {
	mock := newMock(t)
	repo := slotrepo.New(mock)

	slotID := uuid.New()

	rows := pgxmock.NewRows([]string{"id", "room_id", "start_time", "end_time", "created_at"})
	mock.ExpectQuery(`SELECT id, room_id`).
		WithArgs(slotID).
		WillReturnRows(rows)

	result, err := repo.GetByID(context.Background(), slotID)
	require.ErrorIs(t, err, domain.ErrSlotNotFound)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID_Error(t *testing.T) {
	mock := newMock(t)
	repo := slotrepo.New(mock)

	slotID := uuid.New()

	mock.ExpectQuery(`SELECT id, room_id`).
		WithArgs(slotID).
		WillReturnError(assert.AnError)

	result, err := repo.GetByID(context.Background(), slotID)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCountByRoomAndDate_Success(t *testing.T) {
	mock := newMock(t)
	repo := slotrepo.New(mock)

	roomID := uuid.New()
	from := time.Now()
	to := from.Add(24 * time.Hour)

	rows := pgxmock.NewRows([]string{"count"}).AddRow(5)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(roomID, from, to).
		WillReturnRows(rows)

	count, err := repo.CountByRoomAndDate(context.Background(), roomID, from, to)
	require.NoError(t, err)
	assert.Equal(t, 5, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCountByRoomAndDate_Error(t *testing.T) {
	mock := newMock(t)
	repo := slotrepo.New(mock)

	roomID := uuid.New()
	from := time.Now()
	to := from.Add(24 * time.Hour)

	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(roomID, from, to).
		WillReturnError(assert.AnError)

	count, err := repo.CountByRoomAndDate(context.Background(), roomID, from, to)
	require.Error(t, err)
	assert.Equal(t, 0, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListAvailable_Success(t *testing.T) {
	mock := newMock(t)
	repo := slotrepo.New(mock)

	roomID := uuid.New()
	from := time.Now()
	to := from.Add(24 * time.Hour)
	now := time.Now()

	rows := pgxmock.NewRows([]string{"id", "room_id", "start_time", "end_time", "created_at"}).
		AddRow(uuid.New(), roomID, now, now.Add(30*time.Minute), now)

	mock.ExpectQuery(`SELECT s.id`).
		WithArgs(roomID, from, to, domain.BookingStatusActive).
		WillReturnRows(rows)

	result, err := repo.ListAvailable(context.Background(), roomID, from, to)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListAvailable_Error(t *testing.T) {
	mock := newMock(t)
	repo := slotrepo.New(mock)

	roomID := uuid.New()
	from := time.Now()
	to := from.Add(24 * time.Hour)

	mock.ExpectQuery(`SELECT s.id`).
		WithArgs(roomID, from, to, domain.BookingStatusActive).
		WillReturnError(assert.AnError)

	result, err := repo.ListAvailable(context.Background(), roomID, from, to)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBulkUpsert_EmptySlots(t *testing.T) {
	mock := newMock(t)
	repo := slotrepo.New(mock)

	err := repo.BulkUpsert(context.Background(), []domain.Slot{})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBulkUpsert_Success(t *testing.T) {
	mock := newMock(t)
	repo := slotrepo.New(mock)

	slots := []domain.Slot{
		{ID: uuid.New(), RoomID: uuid.New(), StartTime: time.Now(), EndTime: time.Now().Add(30 * time.Minute)},
	}

	mock.ExpectBegin()
	batchExp := mock.ExpectBatch()
	batchExp.ExpectExec(`INSERT INTO slots`).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	err := repo.BulkUpsert(context.Background(), slots)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBulkUpsert_BeginError(t *testing.T) {
	mock := newMock(t)
	repo := slotrepo.New(mock)

	slots := []domain.Slot{
		{ID: uuid.New(), RoomID: uuid.New(), StartTime: time.Now(), EndTime: time.Now().Add(30 * time.Minute)},
	}

	mock.ExpectBegin().WillReturnError(assert.AnError)

	err := repo.BulkUpsert(context.Background(), slots)
	require.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
