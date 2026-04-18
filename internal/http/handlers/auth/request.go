package authhandler

// DummyLoginRequest is the request body for DummyLogin.
type DummyLoginRequest struct {
	Role string `json:"role"`
} //@name DummyLoginRequest

// RegisterRequest is the request body for Register.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
} //@name RegisterRequest

// LoginRequest is the request body for Login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
} //@name LoginRequest
