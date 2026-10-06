package workspace

import (
	"context"
	"fmt"

	"strings"

	"kanban-board-be/internal/feature/membership"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

func (r *Repository) Create(ctx context.Context, workspace *Workspace) error {
	query :=
		`INSERT INTO workspaces (
			name,
			description
		)
	VALUES ($1, $2)
	RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(ctx, query, workspace.Name, workspace.Description).Scan(&workspace.ID, &workspace.CreatedAt, &workspace.UpdatedAt)
}

func (r *Repository) CreateTx(
	ctx context.Context,
	tx pgx.Tx,
	workspace *Workspace,
) error {
	query := `
		INSERT INTO workspaces (
			name,
			description
		)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at
	`

	return tx.QueryRow(
		ctx,
		query,
		workspace.Name,
		workspace.Description,
	).Scan(
		&workspace.ID,
		&workspace.CreatedAt,
		&workspace.UpdatedAt,
	)
}

func (r *Repository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*WorkspaceSummary, error) {
	query := `
		WITH user_workspaces AS (
			SELECT
				w.id,
				w.name,
				w.description,
				wm.role,
				w.created_at,
				w.updated_at
			FROM workspaces w
			INNER JOIN workspace_memberships wm
				ON wm.workspace_id = w.id
			WHERE wm.user_id = $1
		),
		member_counts AS (
			SELECT m.workspace_id, COUNT(*) AS member_count
			FROM workspace_memberships m
			INNER JOIN user_workspaces uw ON uw.id = m.workspace_id
			GROUP BY m.workspace_id
		),
		board_counts AS (
			SELECT b.workspace_id, COUNT(*) AS board_count
			FROM boards b
			INNER JOIN user_workspaces uw ON uw.id = b.workspace_id
			GROUP BY b.workspace_id
		)
		SELECT
			uw.id,
			uw.name,
			uw.description,
			uw.role,
			COALESCE(mc.member_count, 0) AS member_count,
			COALESCE(bc.board_count, 0) AS board_count,
			uw.created_at,
			uw.updated_at
		FROM user_workspaces uw
		LEFT JOIN member_counts mc ON mc.workspace_id = uw.id
		LEFT JOIN board_counts bc ON bc.workspace_id = uw.id
		ORDER BY uw.created_at ASC
	`

	workspaces := []*WorkspaceSummary{}
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query workspaces: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		workspace := &WorkspaceSummary{}
		err := rows.Scan(
			&workspace.ID,
			&workspace.Name,
			&workspace.Description,
			&workspace.Role,
			&workspace.MemberCount,
			&workspace.BoardCount,
			&workspace.CreatedAt,
			&workspace.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}
		workspaces = append(workspaces, workspace)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over workspace rows: %w", err)
	}

	return workspaces, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Workspace, error) {
	query := `
		SELECT
			id,
			name,
			description,
			created_at,
			updated_at
		FROM workspaces
		WHERE id = $1
	`

	workspace := &Workspace{}

	err := r.pool.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&workspace.ID,
		&workspace.Name,
		&workspace.Description,
		&workspace.CreatedAt,
		&workspace.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get workspace by ID: %w", err)
	}

	return workspace, nil
}

func (r *Repository) GetList(ctx context.Context) ([]*Workspace, error) {
	query := `
		SELECT
			id,
			name,
			description,
			created_at,
			updated_at
		FROM workspaces
	`
	workspaces := []*Workspace{}
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query workspaces: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		workspace := &Workspace{}
		err := rows.Scan(
			&workspace.ID,
			&workspace.Name,
			&workspace.Description,
			&workspace.CreatedAt,
			&workspace.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}
		workspaces = append(workspaces, workspace)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over workspace rows: %w", err)
	}

	return workspaces, nil
}

func (r *Repository) GetWorkspaceName(ctx context.Context, workspaceID uuid.UUID) (string, error) {
	query := `
SELECT name
FROM workspaces
WHERE id = $1
`

	var name string
	err := r.pool.QueryRow(ctx, query, workspaceID).Scan(&name)
	if err != nil {
		return "", fmt.Errorf("failed to get workspace name: %w", err)
	}

	return name, nil
}
func (r *Repository) GetMembership(
	ctx context.Context,
	userID, workspaceID uuid.UUID,
) (*membership.WorkspaceMembership, error) {
	query := `
		SELECT
			user_id,
			workspace_id,
			role
		FROM workspace_memberships
		WHERE user_id = $1
		AND workspace_id = $2
	`

	membership := &membership.WorkspaceMembership{}

	err := r.pool.QueryRow(
		ctx,
		query,
		userID,
		workspaceID,
	).Scan(
		&membership.UserID,
		&membership.WorkspaceID,
		&membership.Role,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get workspace membership: %w", err)
	}

	return membership, nil
}

func (r *Repository) UpdateWorkspace(ctx context.Context, update *WorkspaceUpdate) error {
	var setClauses []string
	var args []interface{}

	if update.Name != nil {
		setClauses = append(setClauses, "name = $1")
		args = append(args, *update.Name)
	}

	if update.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", len(args)+1))
		args = append(args, *update.Description)
	}

	if len(setClauses) == 0 {
		return nil // Nothing to update
	}

	args = append(args, update.ID) // Add the workspace ID as the last argument

	query := fmt.Sprintf(`
		UPDATE workspaces
		SET %s,
			updated_at = NOW()
		WHERE id = $%d
	`,
		strings.Join(setClauses, ", "),
		len(args), // The position of the workspace ID in the args slice
	)

	_, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	return nil
}

func (r *Repository) DeleteWorkspace(ctx context.Context, workspaceID uuid.UUID) error {
	query := `
		DELETE FROM workspaces
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}
	return nil
}
