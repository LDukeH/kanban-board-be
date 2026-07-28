package board

import (
	"context"
	"fmt"
	"kanban-board-be/internal/feature/membership"
	"kanban-board-be/internal/feature/workspace"

	"github.com/google/uuid"
)

type Service struct {
	boardRepo     *Repository
	workspaceRepo *workspace.Repository
}

func NewService(boardRepo *Repository, workspaceRepo *workspace.Repository) *Service {
	return &Service{
		boardRepo:     boardRepo,
		workspaceRepo: workspaceRepo,
	}
}

func (s *Service) CreateBoard(ctx context.Context, userID uuid.UUID, board *Board) error {
	// Check if the workspace exists
	_, err := s.workspaceRepo.GetByID(ctx, board.WorkspaceID)
	if err != nil {
		return err
	}

	// Check if the user has permission to create a board in the workspace
	m, err := s.workspaceRepo.GetMembership(ctx, userID, board.WorkspaceID)
	if err != nil {
		return err
	}
	
	if m.Role != membership.Owner && m.Role != membership.Admin {
		return fmt.Errorf("user does not have permission to create a board in this workspace")
	}
	
	// Create the board
	return s.boardRepo.Create(ctx, board)
}

func (s *Service) GetBoardsByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*BoardSummary, error) {
	// Check if the workspace exists
	_, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Get the boards for the workspace
	return s.boardRepo.GetByWorkspaceID(ctx, workspaceID)
}