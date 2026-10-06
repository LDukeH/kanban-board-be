package membership

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (r *Repository) GetWorkspaceRole(ctx context.Context, userID, workspaceID uuid.UUID) (WorkspaceRole, error) {
	query := `
		SELECT role
		FROM workspace_memberships
		WHERE user_id = $1 AND workspace_id = $2
	`

	var role WorkspaceRole
	if err := r.pool.QueryRow(ctx, query, userID, workspaceID).Scan(&role); err != nil {
		return "", fmt.Errorf("failed to get workspace role: %w", err)
	}

	return role, nil
}

func (r *Repository) GetBoardRole(ctx context.Context, userID, boardID uuid.UUID) (uuid.UUID, WorkspaceRole, error) {
	query := `
		SELECT b.workspace_id, wm.role
		FROM boards b
		JOIN workspace_memberships wm ON wm.workspace_id = b.workspace_id
		WHERE b.id = $1 AND wm.user_id = $2
	`

	var workspaceID uuid.UUID
	var role WorkspaceRole
	if err := r.pool.QueryRow(ctx, query, boardID, userID).Scan(&workspaceID, &role); err != nil {
		return uuid.Nil, "", fmt.Errorf("failed to get board role: %w", err)
	}

	return workspaceID, role, nil
}

func (r *Repository) GetColumnRole(ctx context.Context, userID, columnID uuid.UUID) (uuid.UUID, WorkspaceRole, error) {
	query := `
		SELECT c.board_id, wm.role
		FROM columns c
		JOIN boards b ON b.id = c.board_id
		JOIN workspace_memberships wm ON wm.workspace_id = b.workspace_id
		WHERE c.id = $1 AND wm.user_id = $2
	`

	var boardID uuid.UUID
	var role WorkspaceRole
	if err := r.pool.QueryRow(ctx, query, columnID, userID).Scan(&boardID, &role); err != nil {
		return uuid.Nil, "", fmt.Errorf("failed to get column role: %w", err)
	}

	return boardID, role, nil
}

func (r *Repository) GetTaskRole(ctx context.Context, userID, taskID uuid.UUID) (uuid.UUID, uuid.UUID, WorkspaceRole, error) {
	query := `
		SELECT t.column_id, c.board_id, wm.role
		FROM tasks t
		JOIN columns c ON c.id = t.column_id
		JOIN boards b ON b.id = c.board_id
		JOIN workspace_memberships wm ON wm.workspace_id = b.workspace_id
		WHERE t.id = $1 AND wm.user_id = $2
	`

	var columnID uuid.UUID
	var boardID uuid.UUID
	var role WorkspaceRole
	if err := r.pool.QueryRow(ctx, query, taskID, userID).Scan(&columnID, &boardID, &role); err != nil {
		return uuid.Nil, uuid.Nil, "", fmt.Errorf("failed to get task role: %w", err)
	}

	return columnID, boardID, role, nil
}

func (r *Repository) GetLabelRole(ctx context.Context, userID, labelID uuid.UUID) (uuid.UUID, WorkspaceRole, error) {
	query := `
		SELECT l.board_id, wm.role
		FROM labels l
		JOIN boards b ON b.id = l.board_id
		JOIN workspace_memberships wm ON wm.workspace_id = b.workspace_id
		WHERE l.id = $1 AND wm.user_id = $2
	`

	var boardID uuid.UUID
	var role WorkspaceRole
	if err := r.pool.QueryRow(ctx, query, labelID, userID).Scan(&boardID, &role); err != nil {
		return uuid.Nil, "", fmt.Errorf("failed to get label role: %w", err)
	}

	return boardID, role, nil
}
