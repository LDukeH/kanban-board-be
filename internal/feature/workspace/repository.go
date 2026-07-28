package workspace

import (
	"context"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5/pgxpool"

	"fmt"

	"kanban-board-be/internal/feature/membership"
)


type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) Create(ctx context.Context, workspace *Workspace) error {
	query:=
	`INSERT INTO workspaces (
			name
		)
	VALUES ($1)
	RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(ctx, query, workspace.Name).Scan(&workspace.ID, &workspace.CreatedAt, &workspace.UpdatedAt)
}

func (r *Repository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*WorkspaceSummary, error) {
	query := 
	`SELECT
		w.id,
		w.name,
		wm.role,
		w.created_at,
		w.updated_at
	FROM workspaces w
	INNER JOIN workspace_memberships wm
		ON wm.workspace_id = w.id
	WHERE wm.user_id = $1
	ORDER BY w.created_at ASC
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
			&workspace.Role,
			&workspace.CreatedAt,
			&workspace.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}
		workspaces = append(workspaces, workspace)
	}

	return workspaces, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Workspace, error) {
	query := `
		SELECT
			id,
			name,
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
		&workspace.CreatedAt,
		&workspace.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get workspace by ID: %w", err)
	}

	return workspace, nil	
}

func (r *Repository) GetList(ctx context.Context) ([]*Workspace, error) {
	query:=`
		SELECT
			id,
			name,
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
			&workspace.CreatedAt,
			&workspace.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}
		workspaces = append(workspaces, workspace)
	}

	return workspaces, nil
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