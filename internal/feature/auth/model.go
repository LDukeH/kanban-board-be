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

// GetMeResponse is the public shape of `GET /auth/me`. It deliberately omits
// the password hash: the hash is server-side secret material that no client
// has a reason to receive.
type GetMeResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type contextKey string

const UserIDKey contextKey = "user_id"
