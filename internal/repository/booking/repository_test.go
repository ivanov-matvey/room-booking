package booking_test

import (
	"context"
	"testing"
	"time"

	bookingrepo "github.com/ivanov-matvey/room-booking/internal/repository/booking"

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
	repo := bookingrepo.New(mock)
	assert.NotNil(t, repo)
}

func TestCreate_Success(t *testing.T) {
	mock := newMock(t)
	repo := bookingrepo.New(mock)

	now := time.Now()
	booking := &domain.Booking{
		ID:     uuid.New(),
		SlotID: uuid.New(),
		UserID: uuid.New(),
		Status: "active",
	}

	mock.ExpectQuery(`INSERT INTO bookings`).
		WithArgs(booking.ID, booking.SlotID, booking.UserID, booking.Status, booking.ConferenceLink).
		WillReturnRows(pgxmock.NewRows([]string{"created_at"}).AddRow(now))

	err := repo.Create(context.Background(), booking)
	require.NoError(t, err)
	assert.Equal(t, now, booking.CreatedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreate_UniqueViolation(t *testing.T) {
	mock := newMock(t)
	repo := bookingrepo.New(mock)

	booking := &domain.Booking{
		ID:     uuid.New(),
		SlotID: uuid.New(),
		UserID: uuid.New(),
		Status: "active",
	}

	pgErr := &pgconn.PgError{Code: "23505"}
	mock.ExpectQuery(`INSERT INTO bookings`).
		WithArgs(booking.ID, booking.SlotID, booking.UserID, booking.Status, booking.ConferenceLink).
		WillReturnError(pgErr)

	err := repo.Create(context.Background(), booking)
	require.ErrorIs(t, err, domain.ErrSlotBooked)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreate_Error(t *testing.T) {
	mock := newMock(t)
	repo := bookingrepo.New(mock)

	booking := &domain.Booking{
		ID:     uuid.New(),
		SlotID: uuid.New(),
		UserID: uuid.New(),
		Status: "active",
	}

	mock.ExpectQuery(`INSERT INTO bookings`).
		WithArgs(booking.ID, booking.SlotID, booking.UserID, booking.Status, booking.ConferenceLink).
		WillReturnError(assert.AnError)

	err := repo.Create(context.Background(), booking)
	require.Error(t, err)
	require.NotErrorIs(t, err, domain.ErrSlotBooked)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID_Success(t *testing.T) {
	mock := newMock(t)
	repo := bookingrepo.New(mock)

	bookingID := uuid.New()
	now := time.Now()

	rows := pgxmock.NewRows([]string{"id", "slot_id", "user_id", "status", "conference_link", "created_at"}).
		AddRow(bookingID, uuid.New(), uuid.New(), "active", nil, now)

	mock.ExpectQuery(`SELECT id, slot_id`).
		WithArgs(bookingID).
		WillReturnRows(rows)

	result, err := repo.GetByID(context.Background(), bookingID)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, bookingID, result.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID_NotFound(t *testing.T) {
	mock := newMock(t)
	repo := bookingrepo.New(mock)

	bookingID := uuid.New()

	rows := pgxmock.NewRows([]string{"id", "slot_id", "user_id", "status", "conference_link", "created_at"})
	mock.ExpectQuery(`SELECT id, slot_id`).
		WithArgs(bookingID).
		WillReturnRows(rows)

	result, err := repo.GetByID(context.Background(), bookingID)
	require.ErrorIs(t, err, domain.ErrBookingNotFound)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID_Error(t *testing.T) {
	mock := newMock(t)
	repo := bookingrepo.New(mock)

	bookingID := uuid.New()

	mock.ExpectQuery(`SELECT id, slot_id`).
		WithArgs(bookingID).
		WillReturnError(assert.AnError)

	result, err := repo.GetByID(context.Background(), bookingID)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCancel_Success(t *testing.T) {
	mock := newMock(t)
	repo := bookingrepo.New(mock)

	bookingID := uuid.New()

	mock.ExpectExec(`UPDATE bookings`).
		WithArgs(bookingID, domain.BookingStatusCancelled).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err := repo.Cancel(context.Background(), bookingID)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCancel_Error(t *testing.T) {
	mock := newMock(t)
	repo := bookingrepo.New(mock)

	bookingID := uuid.New()

	mock.ExpectExec(`UPDATE bookings`).
		WithArgs(bookingID, domain.BookingStatusCancelled).
		WillReturnError(assert.AnError)

	err := repo.Cancel(context.Background(), bookingID)
	require.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListAll_Success(t *testing.T) {
	mock := newMock(t)
	repo := bookingrepo.New(mock)

	now := time.Now()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM bookings`).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))

	rows := pgxmock.NewRows([]string{"id", "slot_id", "user_id", "status", "conference_link", "created_at"}).
		AddRow(uuid.New(), uuid.New(), uuid.New(), "active", nil, now)
	mock.ExpectQuery(`SELECT id, slot_id, user_id`).
		WithArgs(20, 0).
		WillReturnRows(rows)

	bookings, total, err := repo.ListAll(context.Background(), 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, bookings, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListAll_QueryError(t *testing.T) {
	mock := newMock(t)
	repo := bookingrepo.New(mock)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM bookings`).
		WillReturnError(assert.AnError)

	bookings, total, err := repo.ListAll(context.Background(), 1, 20)
	require.Error(t, err)
	assert.Equal(t, 0, total)
	assert.Nil(t, bookings)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListByUserFuture_Success(t *testing.T) {
	mock := newMock(t)
	repo := bookingrepo.New(mock)

	userID := uuid.New()
	now := time.Now()

	rows := pgxmock.NewRows([]string{"id", "slot_id", "user_id", "status", "conference_link", "created_at"}).
		AddRow(uuid.New(), uuid.New(), userID, "active", nil, now)

	mock.ExpectQuery(`SELECT b.id`).
		WithArgs(userID).
		WillReturnRows(rows)

	result, err := repo.ListByUserFuture(context.Background(), userID)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListByUserFuture_Error(t *testing.T) {
	mock := newMock(t)
	repo := bookingrepo.New(mock)

	userID := uuid.New()

	mock.ExpectQuery(`SELECT b.id`).
		WithArgs(userID).
		WillReturnError(assert.AnError)

	result, err := repo.ListByUserFuture(context.Background(), userID)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}
