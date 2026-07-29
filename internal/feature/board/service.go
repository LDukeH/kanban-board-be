package board

import (
	"context"

	"kanban-board-be/internal/feature/authz"
	"kanban-board-be/internal/feature/membership"

	"github.com/google/uuid"
)

type Service struct {
	boardRepo *Repository
	authz     *authz.Service
}

func NewService(boardRepo *Repository, authz *authz.Service) *Service {
	return &Service{
		boardRepo: boardRepo,
		authz:     authz,
	}
}

func (s *Service) CreateBoard(ctx context.Context, userID uuid.UUID, board *Board) error {
	if err := s.authz.RequireWorkspaceRole(
		ctx,
		userID,
		board.WorkspaceID,
		membership.Owner,
		membership.Admin,
	); err != nil {
		return err
	}

	return s.boardRepo.Create(ctx, board)
}

func (s *Service) GetBoardsByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID) ([]*BoardSummary, error) {
	if err := s.authz.RequireWorkspaceRole(
		ctx,
		userID,
		workspaceID,
		membership.Owner,
		membership.Admin,
		membership.Member,
	); err != nil {
		return nil, err
	}

	return s.boardRepo.GetByWorkspaceID(ctx, workspaceID)
}

func (s *Service) GetBoardByID(ctx context.Context, boardID uuid.UUID, userID uuid.UUID) (*Board, error) {
	board, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return nil, err
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

	return board, nil
}
