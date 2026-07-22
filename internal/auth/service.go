package auth

import (
	"kanban-board-be/internal/user"

	"golang.org/x/crypto/bcrypt"

	"context"
)

type Service struct {
	userService *user.Service
}

func NewService(userService *user.Service) *Service {
	return &Service{
		userService: userService,
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


func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}