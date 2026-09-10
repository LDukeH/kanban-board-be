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

// TaskColumnProvider can resolve a task ID to its column ID.
type TaskColumnProvider interface {
	GetTaskColumnID(ctx context.Context, taskID uuid.UUID) (uuid.UUID, error)
}

// LabelBoardProvider can resolve a label ID to its board ID.
type LabelBoardProvider interface {
	GetLabelBoardID(ctx context.Context, labelID uuid.UUID) (uuid.UUID, error)
}

// Service provides authorization checks for workspace-scoped resources.
type Service struct {
	membershipProvider MembershipProvider
	boardProvider      BoardWorkspaceProvider
	columnProvider     ColumnBoardProvider
	taskProvider       TaskColumnProvider
	labelProvider      LabelBoardProvider
}

// NewService creates a new authorization service.
// Only inject the providers that are actually needed.
func NewService(
	membershipProvider MembershipProvider,
	boardProvider BoardWorkspaceProvider,
	columnProvider ColumnBoardProvider,
	taskProvider TaskColumnProvider,
	labelProvider LabelBoardProvider,
) *Service {
	return &Service{
		membershipProvider: membershipProvider,
		boardProvider:      boardProvider,
		columnProvider:     columnProvider,
		taskProvider:       taskProvider,
		labelProvider:      labelProvider,
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
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperrors.ErrForbidden
		}
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

// GetTaskRole resolves the task's column, board, and workspace, then returns the user's role.
func (s *Service) GetTaskRole(
	ctx context.Context,
	userID uuid.UUID,
	taskID uuid.UUID,
) (membership.WorkspaceRole, error) {
	columnID, err := s.taskProvider.GetTaskColumnID(ctx, taskID)
	if err != nil {
		return "", fmt.Errorf("get task column: %w", err)
	}
	return s.GetColumnRole(ctx, userID, columnID)
}

// GetLabelRole resolves the label's board and workspace, then returns the user's role.
func (s *Service) GetLabelRole(
	ctx context.Context,
	userID uuid.UUID,
	labelID uuid.UUID,
) (membership.WorkspaceRole, error) {
	boardID, err := s.labelProvider.GetLabelBoardID(ctx, labelID)
	if err != nil {
		return "", fmt.Errorf("get label board: %w", err)
	}
	return s.GetBoardRole(ctx, userID, boardID)
}

// GetColumnBoardID returns the board ID for a given column.
func (s *Service) GetColumnBoardID(ctx context.Context, columnID uuid.UUID) (uuid.UUID, error) {
	return s.columnProvider.GetColumnBoardID(ctx, columnID)
}
