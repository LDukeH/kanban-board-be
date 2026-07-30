package auth

import (
	"context"
	"errors"
	"log"
	"time"

	"kanban-board-be/internal/apperrors"
	"kanban-board-be/internal/feature/user"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	userService *user.Service
	config      Config
}

func NewService(userService *user.Service, config Config) *Service {
	return &Service{
		userService: userService,
		config:      config,
	}
}

func (s *Service) RegisterUser(ctx context.Context, email, plainPassword string) (*user.User, error) {
	hash, err := HashPassword(plainPassword)
	if err != nil {
		return nil, err
	}

	newUser := &user.User{
		Email:        email,
		PasswordHash: hash,
	}

	if err := s.userService.CreateUser(ctx, newUser); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, errors.Join(apperrors.ErrConflict, err)
		}
		return nil, err
	}

	return newUser, nil
}

func (s *Service) LookupUser(ctx context.Context) (*user.User, error) {
	userID, err := GetUserIDCtx(ctx)
	if err != nil {
		return nil, err
	}

	return s.userService.GetUserByID(ctx, userID.String())
}

func (s *Service) LoginUser(ctx context.Context, email, plainPassword string) (string, error) {
	user, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errors.Join(apperrors.ErrNotFound, err)
		}
		return "", err
	}

	// Compare the provided password with the stored hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(plainPassword)); err != nil {
		log.Printf("Password mismatch for user %s: %v", email, err)
		return "", errors.Join(apperrors.ErrInvalidCredential, err)
	}

	token, err := GenerateToken(user.ID.String(), s.config.SecretKey)
	if err != nil {
		log.Printf("Failed to generate token for user %s: %v", email, err)
		return "", err
	}

	return token, nil
}


func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func GenerateToken(userID string, secretKey string) (string, error) {
	claims:= jwt.MapClaims{
		"user_id": userID,
		"exp":     jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Token expires in 24 hours
		"iat":     jwt.NewNumericDate(time.Now()), // Issued at
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secretKey))
}

func ValidateToken(tokenString string, secretKey string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		log.Printf("Failed to parse token: %v", err)
		return "", errors.Join(apperrors.ErrUnauthorized, err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok || !token.Valid {
		log.Printf("Invalid token claims: %v", claims)
		return "", errors.Join(apperrors.ErrUnauthorized, jwt.ErrSignatureInvalid)
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		log.Printf("User ID not found in token claims: %v", claims)
		return "", errors.Join(apperrors.ErrUnauthorized, jwt.ErrSignatureInvalid)
	}
	return userID, nil
}
