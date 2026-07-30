package column

import (
	"context"
	"fmt"

	"kanban-board-be/internal/apperrors"
	"kanban-board-be/internal/feature/authz"
	"kanban-board-be/internal/feature/board"
	"kanban-board-be/internal/feature/membership"

	"github.com/google/uuid"
)

type columnRepository interface {
	Create(ctx context.Context, column *Column) error
	GetByBoardID(ctx context.Context, boardID uuid.UUID) ([]*Column, error)
	GetByID(ctx context.Context, columnID string) (*Column, error)
}

type boardRepository interface {
	GetByID(ctx context.Context, boardID uuid.UUID) (*board.Board, error)
}

type Service struct {
	boardRepo  boardRepository
	columnRepo columnRepository
	authz      *authz.Service
}

func NewService(boardRepo boardRepository, columnRepo columnRepository, authz *authz.Service) *Service {
	return &Service{
		boardRepo:  boardRepo,
		columnRepo: columnRepo,
		authz:      authz,
	}
}

func (s *Service) CreateColumn(ctx context.Context, userID uuid.UUID, column *Column) error {
	role, err := s.authz.GetBoardRole(ctx, userID, column.BoardID)
	if err != nil {
		return err
	}
	if role != membership.Owner && role != membership.Admin {
		return apperrors.ErrForbidden
	}
	// position will be handled in repository layer, so we don't need to set it here

	return s.columnRepo.Create(ctx, column)
}

func (s *Service) GetColumnByID(ctx context.Context, columnID string) (*Column, error) {
	return s.columnRepo.GetByID(ctx, columnID)
}

func (s *Service) GetColumnsByBoardID(ctx context.Context, boardID uuid.UUID, userID uuid.UUID) ([]*Column, error) {
	// Check if the board exists
	_, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return nil, fmt.Errorf("board not found: %w", err)
	}

	if _, err := s.authz.GetBoardRole(ctx, userID, boardID); err != nil {
		return nil, err
	}

	return s.columnRepo.GetByBoardID(ctx, boardID)
}
