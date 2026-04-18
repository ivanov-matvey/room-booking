package schedule_test

import (
	"context"
	"testing"
	"time"

	schedulerepo "github.com/ivanov-matvey/room-booking/internal/repository/schedule"

	"github.com/ivanov-matvey/room-booking/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	pgxmock "github.com/pashagolub/pgxmock/v4"
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
	repo := schedulerepo.New(mock)
	assert.NotNil(t, repo)
}

func TestGetByRoomID_NotFound(t *testing.T) {
	mock := newMock(t)
	repo := schedulerepo.New(mock)

	roomID := uuid.New()

	rows := pgxmock.NewRows([]string{"id", "room_id", "days_of_week", "start_time", "end_time", "created_at"})
	mock.ExpectQuery(`SELECT id, room_id`).
		WithArgs(roomID).
		WillReturnRows(rows)

	result, err := repo.GetByRoomID(context.Background(), roomID)
	require.NoError(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByRoomID_Error(t *testing.T) {
	mock := newMock(t)
	repo := schedulerepo.New(mock)

	roomID := uuid.New()

	mock.ExpectQuery(`SELECT id, room_id`).
		WithArgs(roomID).
		WillReturnError(assert.AnError)

	result, err := repo.GetByRoomID(context.Background(), roomID)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateWithSlots_Success(t *testing.T) {
	mock := newMock(t)
	repo := schedulerepo.New(mock)

	schedule := &domain.Schedule{
		ID:         uuid.New(),
		RoomID:     uuid.New(),
		DaysOfWeek: []int{1, 2, 3},
		StartTime:  "09:00",
		EndTime:    "18:00",
	}
	slots := []domain.Slot{
		{ID: uuid.New(), RoomID: schedule.RoomID, StartTime: time.Now(), EndTime: time.Now().Add(30 * time.Minute)},
	}

	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO schedules`).
		WithArgs(schedule.ID, schedule.RoomID, schedule.DaysOfWeek, schedule.StartTime, schedule.EndTime).
		WillReturnRows(pgxmock.NewRows([]string{"created_at"}).AddRow(now))
	batchExp := mock.ExpectBatch()
	batchExp.ExpectExec(`INSERT INTO slots`).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	err := repo.CreateWithSlots(context.Background(), schedule, slots)
	require.NoError(t, err)
	assert.Equal(t, now, schedule.CreatedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateWithSlots_BeginError(t *testing.T) {
	mock := newMock(t)
	repo := schedulerepo.New(mock)

	schedule := &domain.Schedule{ID: uuid.New(), RoomID: uuid.New(), DaysOfWeek: []int{1}}
	var slots []domain.Slot

	mock.ExpectBegin().WillReturnError(assert.AnError)

	err := repo.CreateWithSlots(context.Background(), schedule, slots)
	require.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateWithSlots_InsertScheduleError(t *testing.T) {
	mock := newMock(t)
	repo := schedulerepo.New(mock)

	schedule := &domain.Schedule{
		ID:         uuid.New(),
		RoomID:     uuid.New(),
		DaysOfWeek: []int{1},
		StartTime:  "09:00",
		EndTime:    "18:00",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO schedules`).
		WithArgs(schedule.ID, schedule.RoomID, schedule.DaysOfWeek, schedule.StartTime, schedule.EndTime).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.CreateWithSlots(context.Background(), schedule, nil)
	require.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateWithSlots_UniqueViolation(t *testing.T) {
	mock := newMock(t)
	repo := schedulerepo.New(mock)

	schedule := &domain.Schedule{
		ID:         uuid.New(),
		RoomID:     uuid.New(),
		DaysOfWeek: []int{1},
		StartTime:  "09:00",
		EndTime:    "18:00",
	}

	pgErr := &pgconn.PgError{Code: "23505"}
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO schedules`).
		WithArgs(schedule.ID, schedule.RoomID, schedule.DaysOfWeek, schedule.StartTime, schedule.EndTime).
		WillReturnError(pgErr)
	mock.ExpectRollback()

	err := repo.CreateWithSlots(context.Background(), schedule, nil)
	require.ErrorIs(t, err, domain.ErrScheduleExists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateWithSlots_NoSlots(t *testing.T) {
	mock := newMock(t)
	repo := schedulerepo.New(mock)

	now := time.Now()
	schedule := &domain.Schedule{
		ID:         uuid.New(),
		RoomID:     uuid.New(),
		DaysOfWeek: []int{1},
		StartTime:  "09:00",
		EndTime:    "18:00",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO schedules`).
		WithArgs(schedule.ID, schedule.RoomID, schedule.DaysOfWeek, schedule.StartTime, schedule.EndTime).
		WillReturnRows(pgxmock.NewRows([]string{"created_at"}).AddRow(now))
	mock.ExpectCommit()

	err := repo.CreateWithSlots(context.Background(), schedule, []domain.Slot{})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
