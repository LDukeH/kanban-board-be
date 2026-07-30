package board

import (
	"context"

	"kanban-board-be/internal/apperrors"
	"kanban-board-be/internal/feature/authz"
	"kanban-board-be/internal/feature/membership"

	"github.com/google/uuid"
)

type boardRepository interface {
	Create(ctx context.Context, board *Board) error
	GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*BoardSummary, error)
	GetByID(ctx context.Context, boardID uuid.UUID) (*Board, error)
	UpdateBoard(ctx context.Context, board *Board) error
	DeleteBoard(ctx context.Context, boardID uuid.UUID) error
}

type Service struct {
	boardRepo boardRepository
	authz     *authz.Service
}

func NewService(boardRepo boardRepository, authz *authz.Service) *Service {
	return &Service{
		boardRepo: boardRepo,
		authz:     authz,
	}
}

func (s *Service) CreateBoard(ctx context.Context, userID uuid.UUID, board *Board) error {
	role, err := s.authz.GetWorkspaceRole(ctx, userID, board.WorkspaceID)
	if err != nil {
		return err
	}
	if role != membership.Owner && role != membership.Admin {
		return apperrors.ErrForbidden
	}

	return s.boardRepo.Create(ctx, board)
}

func (s *Service) GetBoardsByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID) ([]*BoardSummary, error) {
	if _, err := s.authz.GetWorkspaceRole(ctx, userID, workspaceID); err != nil {
		return nil, err
	}

	return s.boardRepo.GetByWorkspaceID(ctx, workspaceID)
}

func (s *Service) GetBoardByID(ctx context.Context, boardID uuid.UUID, userID uuid.UUID) (*Board, error) {
	board, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return nil, err
	}

	if _, err := s.authz.GetBoardRole(ctx, userID, boardID); err != nil {
		return nil, err
	}

	return board, nil
}

func (s *Service) UpdateBoard(ctx context.Context, userID uuid.UUID, board *Board) error {
	role, err := s.authz.GetBoardRole(ctx, userID, board.ID)
	if err != nil {
		return err
	}
	if role != membership.Owner && role != membership.Admin {
		return apperrors.ErrForbidden
	}

	return s.boardRepo.UpdateBoard(ctx, board)
}

func (s *Service) DeleteBoard(ctx context.Context, userID uuid.UUID, boardID uuid.UUID) error {
	role, err := s.authz.GetBoardRole(ctx, userID, boardID)
	if err != nil {
		return err
	}
	if role != membership.Owner && role != membership.Admin {
		return apperrors.ErrForbidden
	}

	return s.boardRepo.DeleteBoard(ctx, boardID)
}