package auth

import (
	"kanban-board-be/internal/user"

	"golang.org/x/crypto/bcrypt"

	"context"

	"github.com/golang-jwt/jwt/v5"
	
	

	"time"

	"log"
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

func (s *Service) RegisterUser(email, plainPassword string) (*user.User, error) {
	hash, err := HashPassword(plainPassword)
	
	if err != nil {
		return nil, err
	}

	newUser := &user.User{
		Email:        email,
		PasswordHash: hash,
	}
	
	return newUser, s.userService.CreateUser(context.Background(), newUser)
}

func (s *Service) LoginUser(ctx context.Context ,email, plainPassword string) (string, error) {
		user, err := s.userService.GetUserByEmail(ctx, email)

		if err !=nil{
			return "", err
		}

	// Compare the provided password with the stored hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(plainPassword)); err != nil {
		log.Printf("Password mismatch for user %s: %v", email, err)
		return "", err // Passwords do not match
	}

	if err != nil {
		log.Printf("Invalid credentials for user %s: %v", email, err)
		return "",err
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