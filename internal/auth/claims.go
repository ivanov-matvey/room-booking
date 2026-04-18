package auth

import "github.com/golang-jwt/jwt/v5"

const (
	AdminUserID   = "00000000-0000-0000-0000-000000000001"
	DefaultUserID = "00000000-0000-0000-0000-000000000002"
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}
