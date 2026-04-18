package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/ivanov-matvey/room-booking/internal/auth"
	"github.com/ivanov-matvey/room-booking/internal/domain"
	authusecase "github.com/ivanov-matvey/room-booking/internal/usecase/auth"
	mockauth "github.com/ivanov-matvey/room-booking/mocks/usecase/auth"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func parseToken(t *testing.T, tokenStr, secret string) *auth.Claims {
	t.Helper()
	claims := &auth.Claims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(tok *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	require.NoError(t, err)
	return claims
}

func TestDummyLogin_Admin(t *testing.T) {
	repo := mockauth.NewMockUserRepository(t)
	uc := authusecase.New(repo, "test-secret", 24*time.Hour)

	token, err := uc.DummyLogin(context.Background(), "admin")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims := parseToken(t, token, "test-secret")
	assert.Equal(t, auth.AdminUserID, claims.UserID)
	assert.Equal(t, "admin", claims.Role)
}

func TestDummyLogin_User(t *testing.T) {
	repo := mockauth.NewMockUserRepository(t)
	uc := authusecase.New(repo, "test-secret", 24*time.Hour)

	token, err := uc.DummyLogin(context.Background(), "user")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims := parseToken(t, token, "test-secret")
	assert.Equal(t, auth.DefaultUserID, claims.UserID)
	assert.Equal(t, "user", claims.Role)
}

func TestRegister_Success(t *testing.T) {
	repo := mockauth.NewMockUserRepository(t)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	uc := authusecase.New(repo, "test-secret", 24*time.Hour)
	user, err := uc.Register(context.Background(), "test@example.com", "password123", "user")
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "user", user.Role)
	assert.NotEqual(t, uuid.Nil, user.ID)
	repo.AssertExpectations(t)
}

func TestRegister_RepoError(t *testing.T) {
	repo := mockauth.NewMockUserRepository(t)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(assert.AnError)

	uc := authusecase.New(repo, "test-secret", 24*time.Hour)
	_, err := uc.Register(context.Background(), "test@example.com", "password123", "user")
	require.Error(t, err)
	repo.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	require.NoError(t, err)
	userID := uuid.New()
	dbUser := &domain.User{
		ID:           userID,
		Email:        "test@example.com",
		Role:         "user",
		PasswordHash: new(string(hash)),
	}

	repo := mockauth.NewMockUserRepository(t)
	repo.On("GetByEmail", mock.Anything, "test@example.com").Return(dbUser, nil)

	uc := authusecase.New(repo, "test-secret", 24*time.Hour)
	token, err := uc.Login(context.Background(), "test@example.com", "password123")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims := parseToken(t, token, "test-secret")
	assert.Equal(t, userID.String(), claims.UserID)
	assert.Equal(t, "user", claims.Role)
	repo.AssertExpectations(t)
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := mockauth.NewMockUserRepository(t)
	repo.On("GetByEmail", mock.Anything, "notfound@example.com").Return(nil, domain.ErrUnauthorized)

	uc := authusecase.New(repo, "test-secret", 24*time.Hour)
	_, err := uc.Login(context.Background(), "notfound@example.com", "password")
	require.ErrorIs(t, err, domain.ErrUnauthorized)
	repo.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.MinCost)
	require.NoError(t, err)
	userID := uuid.New()
	dbUser := &domain.User{
		ID:           userID,
		Email:        "test@example.com",
		Role:         "user",
		PasswordHash: new(string(hash)),
	}

	repo := mockauth.NewMockUserRepository(t)
	repo.On("GetByEmail", mock.Anything, "test@example.com").Return(dbUser, nil)

	uc := authusecase.New(repo, "test-secret", 24*time.Hour)
	_, err = uc.Login(context.Background(), "test@example.com", "wrongpassword")
	require.ErrorIs(t, err, domain.ErrUnauthorized)
	repo.AssertExpectations(t)
}

func TestLogin_NoPasswordHash(t *testing.T) {
	userID := uuid.New()
	dbUser := &domain.User{
		ID:           userID,
		Email:        "admin@example.com",
		Role:         "admin",
		PasswordHash: nil,
	}

	repo := mockauth.NewMockUserRepository(t)
	repo.On("GetByEmail", mock.Anything, "admin@example.com").Return(dbUser, nil)

	uc := authusecase.New(repo, "test-secret", 24*time.Hour)
	_, err := uc.Login(context.Background(), "admin@example.com", "anypassword")
	require.ErrorIs(t, err, domain.ErrUnauthorized)
	repo.AssertExpectations(t)
}
