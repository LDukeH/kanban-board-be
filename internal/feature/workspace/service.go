package workspace

import (
	"context"
	"fmt"

	"kanban-board-be/internal/apperrors"
	"kanban-board-be/internal/feature/authz"
	"kanban-board-be/internal/feature/membership"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type workspaceRepository interface {
	BeginTx(ctx context.Context) (pgx.Tx, error)
	Create(ctx context.Context, workspace *Workspace) error
	CreateTx(ctx context.Context, tx pgx.Tx, workspace *Workspace) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*WorkspaceSummary, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Workspace, error)
	UpdateWorkspace(ctx context.Context, update *WorkspaceUpdate) error
	DeleteWorkspace(ctx context.Context, workspaceID uuid.UUID) error
}

type membershipRepository interface {
	CreateTx(ctx context.Context, tx pgx.Tx, membership *membership.WorkspaceMembership) error
	TransferOwnership(ctx context.Context, oldOwnerID, newOwnerID, workspaceID uuid.UUID) error
}

type Service struct {
	workspaceRepo  workspaceRepository
	membershipRepo membershipRepository
	authz          *authz.Service
}

func NewService(workspaceRepo workspaceRepository, membershipRepo membershipRepository, authz *authz.Service) *Service {
	return &Service{
		workspaceRepo:  workspaceRepo,
		membershipRepo: membershipRepo,
		authz:          authz,
	}
}

// Empty is valid; "" is how a description-less workspace is stored.
func validateWorkspaceDescription(description string) error {
	if len(description) > maxWorkspaceDescriptionLength {
		return apperrors.ErrValidation
	}
	return nil
}

func (s *Service) CreateWorkspace(ctx context.Context, workspace *Workspace, userID uuid.UUID) error {
	if err := validateWorkspaceDescription(workspace.Description); err != nil {
		return err
	}

	tx, err := s.workspaceRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer tx.Rollback(ctx) // Rollback the transaction if not committed

	// Create the workspace
	if err := s.workspaceRepo.CreateTx(ctx, tx, workspace); err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	membership := &membership.WorkspaceMembership{
		WorkspaceID: workspace.ID,
		UserID:      userID,
		Role:        membership.Owner,
	}

	if err := s.membershipRepo.CreateTx(ctx, tx, membership); err != nil {
		return fmt.Errorf("failed to create workspace membership: %w", err)
	}

	return tx.Commit(ctx)
}

// by user id, get all workspaces that the user is a member of
func (s *Service) GetAllWorkspaces(ctx context.Context, userID uuid.UUID) ([]*WorkspaceSummary, error) {
	workspaces, err := s.workspaceRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspaces by user ID: %w", err)
	}
	return workspaces, nil
}

func (s *Service) GetWorkspaceByID(ctx context.Context, id string, userID uuid.UUID) (*Workspace, error) {
	workspaceID, err := uuid.Parse(id)

	if err != nil {
		return nil, fmt.Errorf("invalid workspace ID format: %w", err)
	}
	// Check if the workspace exists
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace by ID: %w", err)
	}

	// Check that the user is a member of the workspace.
	if _, err := s.authz.GetWorkspaceRole(ctx, userID, workspaceID); err != nil {
		return nil, fmt.Errorf("user is not authorized to access this workspace: %w", err)
	}

	return workspace, nil
}

// The owner check runs before the write, so a non-owner cannot bypass it with an empty payload.
func (s *Service) UpdateWorkspace(ctx context.Context, update *WorkspaceUpdate, userID uuid.UUID) error {
	// Check that the user is a owner of the workspace.
	role, err := s.authz.GetWorkspaceRole(ctx, userID, update.ID)

	if err != nil {
		return fmt.Errorf("failed to get workspace role: %w", err)
	}
	if role != membership.Owner {
		return fmt.Errorf("user is not authorized to update this workspace")
	}

	if update.Description != nil {
		if err := validateWorkspaceDescription(*update.Description); err != nil {
			return err
		}
	}

	// Update the workspace
	if err := s.workspaceRepo.UpdateWorkspace(ctx, update); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	return nil
}

func (s *Service) DeleteWorkspace(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID) error {
	// Check that the user is a owner of the workspace.
	role, err := s.authz.GetWorkspaceRole(ctx, userID, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace role: %w", err)
	}
	if role != membership.Owner {
		return fmt.Errorf("user is not authorized to delete this workspace")
	}

	// Delete the workspace
	if err := s.workspaceRepo.DeleteWorkspace(ctx, workspaceID); err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}

	return nil
}

func (s *Service) TransferOwnership(ctx context.Context, workspaceID, oldOwnerID, newOwnerID uuid.UUID) error {
	// Check that the old owner is indeed the current owner of the workspace.
	role, err := s.authz.GetWorkspaceRole(ctx, oldOwnerID, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace role: %w", err)
	}
	if role != membership.Owner {
		return fmt.Errorf("user is not authorized to transfer ownership of this workspace")
	}

	// Transfer ownership
	if err := s.membershipRepo.TransferOwnership(ctx, oldOwnerID, newOwnerID, workspaceID); err != nil {
		return fmt.Errorf("failed to transfer ownership: %w", err)
	}

	return nil
}
