package internal

import "time"

// User is a registered account.
type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

// RegisterRequest is the payload for POST /api/register.
type RegisterRequest struct {
	Username string `json:"username" example:"alice"`
	Password string `json:"password" example:"s3cret123"`
}

// LoginRequest is the payload for POST /api/login.
type LoginRequest struct {
	Username string `json:"username" example:"alice"`
	Password string `json:"password" example:"s3cret123"`
}

// TokenResponse is returned on successful login.
type TokenResponse struct {
	Token string `json:"token"`
}

// ErrorResponse is a generic error payload.
type ErrorResponse struct {
	Error string `json:"error" example:"invalid credentials"`
}
