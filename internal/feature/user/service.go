package user

import (
	"context"

	"fmt"

	"log"

	"github.com/google/uuid"
)

type userRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	SearchByEmail(ctx context.Context, email string) ([]*UserResponse, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type Service struct {
	repo userRepository
}

func NewService(repo userRepository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) CreateUser(ctx context.Context, user *User) error {
	return s.repo.Create(ctx, user)
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

func (s *Service) SearchUsersByEmail(ctx context.Context, email string) ([]*UserResponse, error) {
	return s.repo.SearchByEmail(ctx, email)
}

func (s *Service) DeleteUserByID(ctx context.Context, id string) error {
	userID, err := uuid.Parse(id)
	if err != nil {
		log.Printf("Invalid UUID format: %v", err)
		return fmt.Errorf("Invalid UUID format: %w", err)
	}

	return s.repo.DeleteByID(ctx, userID)
}
