package authhandler

import (
	"time"

	"github.com/google/uuid"
)

// TokenResponse is the response body containing a JWT token.
type TokenResponse struct {
	Token string `json:"token"`
} //@name Token

// UserResponse is the response body for a user.
type UserResponse struct {
	ID        uuid.UUID  `json:"id" swaggertype:"string" format:"uuid"`
	Email     string     `json:"email" format:"email"`
	Role      string     `json:"role" enums:"admin,user"`
	CreatedAt *time.Time `json:"createdAt,omitempty" swaggertype:"string" format:"date-time"`
} //@name User

// RegisterResponse is the response body for Register.
type RegisterResponse struct {
	User *UserResponse `json:"user"`
} //@name RegisterResponse
