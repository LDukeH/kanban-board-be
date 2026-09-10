package auth

import "time"

type Config struct {
	SecretKey  string
	Expiration time.Duration
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type GetMeResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type contextKey string

const UserIDKey contextKey = "user_id"
