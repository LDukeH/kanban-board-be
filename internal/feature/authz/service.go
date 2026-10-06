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

type Service struct {
	resolver Resolver
}

func NewService(resolver Resolver) *Service {
	return &Service{resolver: resolver}
}

func mapAccessError(err error, operation string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return apperrors.ErrForbidden
	}
	return fmt.Errorf("%s: %w", operation, err)
}

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

func (s *Service) GetColumnBoardID(ctx context.Context, userID uuid.UUID, columnID uuid.UUID) (uuid.UUID, error) {
	boardID, _, err := s.resolver.GetColumnRole(ctx, userID, columnID)
	if err != nil {
		return uuid.Nil, mapAccessError(err, "resolve column access")
	}
	return boardID, nil
}
