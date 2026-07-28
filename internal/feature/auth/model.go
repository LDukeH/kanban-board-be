package auth

import "time"

type Config struct {
    SecretKey string
    Expiration time.Duration
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type contextKey string

const userIDKey contextKey = "user_id"