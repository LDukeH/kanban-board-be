package authz

import (
	"context"
	"fmt"

	"kanban-board-be/internal/feature/membership"

	"github.com/google/uuid"
)

// MembershipProvider can load a workspace membership for a user.
type MembershipProvider interface {
	GetMembership(ctx context.Context, userID, workspaceID uuid.UUID) (*membership.WorkspaceMembership, error)
}

// BoardWorkspaceProvider can resolve a board ID to its workspace ID.
type BoardWorkspaceProvider interface {
	GetBoardWorkspaceID(ctx context.Context, boardID uuid.UUID) (uuid.UUID, error)
}

// ColumnBoardProvider can resolve a column ID to its board ID.
type ColumnBoardProvider interface {
	GetColumnBoardID(ctx context.Context, columnID uuid.UUID) (uuid.UUID, error)
}

// Service provides authorization checks for workspace-scoped resources.
type Service struct {
	membershipProvider MembershipProvider
	boardProvider      BoardWorkspaceProvider
	columnProvider     ColumnBoardProvider
}

// NewService creates a new authorization service.
// Only inject the providers that are actually needed.
func NewService(
	membershipProvider MembershipProvider,
	boardProvider BoardWorkspaceProvider,
	columnProvider ColumnBoardProvider,
) *Service {
	return &Service{
		membershipProvider: membershipProvider,
		boardProvider:      boardProvider,
		columnProvider:     columnProvider,
	}
}

// GetWorkspaceRole returns the user's role in the workspace.
func (s *Service) GetWorkspaceRole(
	ctx context.Context,
	userID uuid.UUID,
	workspaceID uuid.UUID,
) (membership.WorkspaceRole, error) {
	m, err := s.membershipProvider.GetMembership(ctx, userID, workspaceID)
	if err != nil {
		return "", fmt.Errorf("get membership: %w", err)
	}

	return m.Role, nil
}

// GetBoardRole resolves the board's workspace and returns the user's role.
func (s *Service) GetBoardRole(
	ctx context.Context,
	userID uuid.UUID,
	boardID uuid.UUID,
) (membership.WorkspaceRole, error) {
	workspaceID, err := s.boardProvider.GetBoardWorkspaceID(ctx, boardID)
	if err != nil {
		return "", fmt.Errorf("get board workspace: %w", err)
	}

	return s.GetWorkspaceRole(ctx, userID, workspaceID)
}

// GetColumnRole resolves the column's board and workspace, then returns the user's role.
func (s *Service) GetColumnRole(
	ctx context.Context,
	userID uuid.UUID,
	columnID uuid.UUID,
) (membership.WorkspaceRole, error) {
	boardID, err := s.columnProvider.GetColumnBoardID(ctx, columnID)
	if err != nil {
		return "", fmt.Errorf("get column board: %w", err)
	}

	return s.GetBoardRole(ctx, userID, boardID)
}
