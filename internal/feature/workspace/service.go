package workspace

import (
	"context"

	"kanban-board-be/internal/feature/membership"

	"github.com/google/uuid"

	"fmt"

)

type Service struct {
	workspaceRepo  *Repository
	membershipRepo *membership.Repository
}

func NewService(workspaceRepo *Repository, membershipRepo *membership.Repository) *Service {
	return &Service{
		workspaceRepo:  workspaceRepo,
		membershipRepo: membershipRepo,
	}
}

func (s *Service) CreateWorkspace(ctx context.Context, workspace *Workspace, userID uuid.UUID) error {
	// Create the workspace
	if err := s.workspaceRepo.Create(ctx, workspace); err != nil {
		return err
	}

	// Create the workspace membership for the user
	membership := &membership.WorkspaceMembership{
		WorkspaceID: workspace.ID,
		UserID:      userID,
		Role:        membership.Owner, // Assign the owner role to the user creating the workspace
	}

	// TODO:
	// Wrap workspace creation + owner membership creation
	// in a single database transaction. So that if one fails, the other won't be created either.

	return s.membershipRepo.Create(ctx, membership)
}

// by user id, get all workspaces that the user is a member of
func (s *Service) GetAllWorkspaces(ctx context.Context, userID uuid.UUID) ([]*WorkspaceSummary, error) {
	workspaces, err := s.workspaceRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspaces by user ID: %w", err)
	}
	return workspaces, nil
}

func (s *Service) GetWorkspaceByID(ctx context.Context, id string) (*Workspace, error) {
	workspaceID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid workspace ID format: %w", err)
	}

	return s.workspaceRepo.GetByID(ctx, workspaceID)
}
