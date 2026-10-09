package auth

type Config struct {
	JWKSURL string
	Issuer  string
}

type GetMeResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type contextKey string

const UserIDKey contextKey = "user_id"
const UserEmailKey contextKey = "user_email"
