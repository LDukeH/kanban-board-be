package user

import (
	"context"

	"fmt"

	"log"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}


func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) CreateUser(ctx context.Context, user *User) error {
	return s.repo.CreateUser(ctx, user)
}

func (s *Service) GetUserByID(ctx context.Context, id string) (*User, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		log.Printf("Invalid UUID format: %v", err)
		return nil, fmt.Errorf("Invalid UUID format: %w", err)
	}

	return s.repo.GetByID(ctx, userID)
}

func (s *Service) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return s.repo.GetByEmail(ctx, email)
}