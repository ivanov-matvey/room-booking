package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/ivanov-matvey/room-booking/internal/auth"
	"github.com/ivanov-matvey/room-booking/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}

type UseCase struct {
	userRepository UserRepository
	jwtSecret      string
	jwtExpiration  time.Duration
}

func New(userRepository UserRepository, jwtSecret string, jwtExpiration time.Duration) *UseCase {
	return &UseCase{
		userRepository: userRepository,
		jwtSecret:      jwtSecret,
		jwtExpiration:  jwtExpiration,
	}
}

func (uc *UseCase) DummyLogin(_ context.Context, role string) (string, error) {
	userID := auth.DefaultUserID
	if role == domain.RoleAdmin {
		userID = auth.AdminUserID
	}

	return uc.generateToken(userID, role)
}

func (uc *UseCase) Register(ctx context.Context, email, password, role string) (*domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		Role:         role,
		PasswordHash: new(string(hash)),
	}

	if err := uc.userRepository.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (uc *UseCase) Login(ctx context.Context, email, password string) (string, error) {
	user, err := uc.userRepository.GetByEmail(ctx, email)
	if err != nil {
		return "", domain.ErrUnauthorized
	}

	if user.PasswordHash == nil {
		return "", domain.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return "", domain.ErrUnauthorized
	}

	return uc.generateToken(user.ID.String(), user.Role)
}

func (uc *UseCase) generateToken(userID, role string) (string, error) {
	claims := auth.Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(uc.jwtExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uc.jwtSecret))
}
