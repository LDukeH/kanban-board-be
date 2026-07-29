package column

import (
	"context"
	"fmt"

	"kanban-board-be/internal/feature/authz"
	"kanban-board-be/internal/feature/board"
	"kanban-board-be/internal/feature/membership"

	"github.com/google/uuid"
)

type Service struct {
	boardRepo  *board.Repository
	columnRepo *Repository
	authz      *authz.Service
}

func NewService(boardRepo *board.Repository, columnRepo *Repository, authz *authz.Service) *Service {
	return &Service{
		boardRepo:  boardRepo,
		columnRepo: columnRepo,
		authz:      authz,
	}
}

func (s *Service) CreateColumn(ctx context.Context, userID uuid.UUID, column *Column) error {
	if err := s.authz.RequireBoardRole(
		ctx,
		userID,
		column.BoardID,
		membership.Owner,
		membership.Admin,
	); err != nil {
		return err
	}

	// Auto-assign the next position
	columns, err := s.columnRepo.GetByBoardID(ctx, column.BoardID)
	if err != nil {
		return fmt.Errorf("get columns for position: %w", err)
	}
	column.Position = len(columns)

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

	if err := s.authz.RequireBoardRole(
		ctx,
		userID,
		boardID,
		membership.Owner,
		membership.Admin,
		membership.Member,
	); err != nil {
		return nil, err
	}

	return s.columnRepo.GetByBoardID(ctx, boardID)
}