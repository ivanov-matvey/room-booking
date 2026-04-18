package user_test

import (
	"context"
	"testing"
	"time"

	userrepo "github.com/ivanov-matvey/room-booking/internal/repository/user"

	"github.com/ivanov-matvey/room-booking/internal/domain"

	"github.com/google/uuid"
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
	repo := userrepo.New(mock)
	assert.NotNil(t, repo)
}

func TestCreate_Success(t *testing.T) {
	mock := newMock(t)
	repo := userrepo.New(mock)

	now := time.Now()
	user := &domain.User{ID: uuid.New(), Email: "test@example.com", Role: "user"}

	rows := pgxmock.NewRows([]string{"created_at"}).AddRow(now)
	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(user.ID, user.Email, user.Role, user.PasswordHash).
		WillReturnRows(rows)

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)
	assert.Equal(t, now, user.CreatedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreate_Error(t *testing.T) {
	mock := newMock(t)
	repo := userrepo.New(mock)

	user := &domain.User{ID: uuid.New(), Email: "test@example.com", Role: "user"}

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(user.ID, user.Email, user.Role, user.PasswordHash).
		WillReturnError(assert.AnError)

	err := repo.Create(context.Background(), user)
	require.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSeedDefaultUsers_Success(t *testing.T) {
	mock := newMock(t)
	repo := userrepo.New(mock)

	mock.ExpectExec(`INSERT INTO users`).
		WillReturnResult(pgxmock.NewResult("INSERT", 2))

	err := repo.SeedDefaultUsers(context.Background())
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSeedDefaultUsers_Error(t *testing.T) {
	mock := newMock(t)
	repo := userrepo.New(mock)

	mock.ExpectExec(`INSERT INTO users`).
		WillReturnError(assert.AnError)

	err := repo.SeedDefaultUsers(context.Background())
	require.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByEmail_Success(t *testing.T) {
	mock := newMock(t)
	repo := userrepo.New(mock)

	userID := uuid.New()
	now := time.Now()
	email := "test@example.com"

	rows := pgxmock.NewRows([]string{"id", "email", "role", "password_hash", "created_at"}).
		AddRow(userID, email, "user", nil, now)

	mock.ExpectQuery(`SELECT id, email`).
		WithArgs(email).
		WillReturnRows(rows)

	result, err := repo.GetByEmail(context.Background(), email)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, email, result.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByEmail_NotFound(t *testing.T) {
	mock := newMock(t)
	repo := userrepo.New(mock)

	rows := pgxmock.NewRows([]string{"id", "email", "role", "password_hash", "created_at"})
	mock.ExpectQuery(`SELECT id, email`).
		WithArgs("notfound@example.com").
		WillReturnRows(rows)

	result, err := repo.GetByEmail(context.Background(), "notfound@example.com")
	require.ErrorIs(t, err, domain.ErrUserNotFound)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByEmail_Error(t *testing.T) {
	mock := newMock(t)
	repo := userrepo.New(mock)

	mock.ExpectQuery(`SELECT id, email`).
		WithArgs("test@example.com").
		WillReturnError(assert.AnError)

	result, err := repo.GetByEmail(context.Background(), "test@example.com")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID_Success(t *testing.T) {
	mock := newMock(t)
	repo := userrepo.New(mock)

	userID := uuid.New()
	now := time.Now()

	rows := pgxmock.NewRows([]string{"id", "email", "role", "password_hash", "created_at"}).
		AddRow(userID, "test@example.com", "user", nil, now)

	mock.ExpectQuery(`SELECT id, email`).
		WithArgs(userID).
		WillReturnRows(rows)

	result, err := repo.GetByID(context.Background(), userID)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID_NotFound(t *testing.T) {
	mock := newMock(t)
	repo := userrepo.New(mock)

	userID := uuid.New()

	rows := pgxmock.NewRows([]string{"id", "email", "role", "password_hash", "created_at"})
	mock.ExpectQuery(`SELECT id, email`).
		WithArgs(userID).
		WillReturnRows(rows)

	result, err := repo.GetByID(context.Background(), userID)
	require.ErrorIs(t, err, domain.ErrUserNotFound)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID_Error(t *testing.T) {
	mock := newMock(t)
	repo := userrepo.New(mock)

	userID := uuid.New()

	mock.ExpectQuery(`SELECT id, email`).
		WithArgs(userID).
		WillReturnError(assert.AnError)

	result, err := repo.GetByID(context.Background(), userID)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}
