package authz

import (
	"context"
	"errors"
	"fmt"

	"kanban-board-be/internal/apperrors"
	"kanban-board-be/internal/feature/membership"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Resolver resolves a user's access to a workspace-scoped resource.
//
// Each method answers "which workspace owns this resource, and what is the
// user's role in it?" with one SQL statement that joins the resource up to
// workspace_memberships. The previous implementation walked the ownership
// chain one hop per query (task -> column -> board -> workspace -> membership),
// costing up to four database round trips per authorized request; on a
// serverless deployment talking to a remote database those round trips, not
// the query plans, dominate response time.
//
// A missing resource and a resource the user cannot see are reported the same
// way (pgx.ErrNoRows) because the two are indistinguishable through this
// interface; the caller maps both to apperrors.ErrForbidden.
type Resolver interface {
	// GetWorkspaceRole returns the user's role in the workspace.
	GetWorkspaceRole(ctx context.Context, userID, workspaceID uuid.UUID) (membership.WorkspaceRole, error)

	// GetBoardRole returns the workspace that owns the board and the user's role in it.
	GetBoardRole(ctx context.Context, userID, boardID uuid.UUID) (uuid.UUID, membership.WorkspaceRole, error)

	// GetColumnRole returns the board that owns the column and the user's role in the column's workspace.
	GetColumnRole(ctx context.Context, userID, columnID uuid.UUID) (uuid.UUID, membership.WorkspaceRole, error)

	// GetTaskRole returns the task's column, that column's board, and the user's role in the task's workspace.
	GetTaskRole(ctx context.Context, userID, taskID uuid.UUID) (columnID, boardID uuid.UUID, role membership.WorkspaceRole, err error)

	// GetLabelRole returns the board that owns the label and the user's role in the label's workspace.
	GetLabelRole(ctx context.Context, userID, labelID uuid.UUID) (uuid.UUID, membership.WorkspaceRole, error)
}

// Service provides authorization checks for workspace-scoped resources.
type Service struct {
	resolver Resolver
}

// NewService creates a new authorization service.
func NewService(resolver Resolver) *Service {
	return &Service{resolver: resolver}
}

// mapAccessError keeps "not a member" and "does not exist" as 403 Forbidden
// (the workspace is the authorization boundary, so a caller never learns which
// of the two happened) while any other failure stays a 500.
func mapAccessError(err error, operation string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return apperrors.ErrForbidden
	}
	return fmt.Errorf("%s: %w", operation, err)
}

// GetWorkspaceRole returns the user's role in the workspace, or ErrForbidden
// when the user is not a member of it.
func (s *Service) GetWorkspaceRole(
	ctx context.Context,
	userID uuid.UUID,
	workspaceID uuid.UUID,
) (membership.WorkspaceRole, error) {
	role, err := s.resolver.GetWorkspaceRole(ctx, userID, workspaceID)
	if err != nil {
		return "", mapAccessError(err, "resolve workspace access")
	}
	return role, nil
}

// GetBoardRole returns the user's role in the workspace that owns the board.
func (s *Service) GetBoardRole(
	ctx context.Context,
	userID uuid.UUID,
	boardID uuid.UUID,
) (membership.WorkspaceRole, error) {
	_, role, err := s.resolver.GetBoardRole(ctx, userID, boardID)
	if err != nil {
		return "", mapAccessError(err, "resolve board access")
	}
	return role, nil
}

// GetColumnRole returns the user's role in the workspace that owns the column.
func (s *Service) GetColumnRole(
	ctx context.Context,
	userID uuid.UUID,
	columnID uuid.UUID,
) (membership.WorkspaceRole, error) {
	_, role, err := s.resolver.GetColumnRole(ctx, userID, columnID)
	if err != nil {
		return "", mapAccessError(err, "resolve column access")
	}
	return role, nil
}

// GetTaskRole returns the user's role in the workspace that owns the task.
func (s *Service) GetTaskRole(
	ctx context.Context,
	userID uuid.UUID,
	taskID uuid.UUID,
) (membership.WorkspaceRole, error) {
	_, _, role, err := s.resolver.GetTaskRole(ctx, userID, taskID)
	if err != nil {
		return "", mapAccessError(err, "resolve task access")
	}
	return role, nil
}

// GetLabelRole returns the user's role in the workspace that owns the label.
func (s *Service) GetLabelRole(
	ctx context.Context,
	userID uuid.UUID,
	labelID uuid.UUID,
) (membership.WorkspaceRole, error) {
	_, role, err := s.resolver.GetLabelRole(ctx, userID, labelID)
	if err != nil {
		return "", mapAccessError(err, "resolve label access")
	}
	return role, nil
}

// GetColumnBoardID returns the board that owns the column after verifying the
// user can access it, so a caller cannot use it to probe another workspace.
func (s *Service) GetColumnBoardID(ctx context.Context, userID uuid.UUID, columnID uuid.UUID) (uuid.UUID, error) {
	boardID, _, err := s.resolver.GetColumnRole(ctx, userID, columnID)
	if err != nil {
		return uuid.Nil, mapAccessError(err, "resolve column access")
	}
	return boardID, nil
}
