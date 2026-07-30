package membership

import (
	"context"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"fmt"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) Create(ctx context.Context, membership *WorkspaceMembership) error {
	query:=
	`INSERT INTO workspace_memberships (
			workspace_id,
			user_id,
			role
		)
	VALUES ($1, $2, $3)
	RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(ctx,query,membership.WorkspaceID,membership.UserID,membership.Role).Scan(&membership.ID,&membership.CreatedAt,&membership.UpdatedAt);
}

func (r *Repository) CreateTx(
    ctx context.Context,
    tx pgx.Tx,
    membership *WorkspaceMembership,
) error {
   query:=
	`INSERT INTO workspace_memberships (
			workspace_id,
			user_id,
			role
		)
	VALUES ($1, $2, $3)
	RETURNING id, created_at, updated_at`
	
	return tx.QueryRow(
		ctx,
		query,
		membership.WorkspaceID,
		membership.UserID,
		membership.Role,
	).Scan(
		&membership.ID,
		&membership.CreatedAt,
		&membership.UpdatedAt,
	)
}

func (r *Repository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*WorkspaceMembership, error) {
	query := `
		SELECT
			id,
			workspace_id,
			user_id,
			role,
			created_at,
			updated_at
		FROM workspace_memberships
		WHERE id = $1
	`

	membership := &WorkspaceMembership{}

	err := r.pool.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&membership.ID,
		&membership.WorkspaceID,
		&membership.UserID,
		&membership.Role,
		&membership.CreatedAt,
		&membership.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get workspace membership by ID: %w", err)
	}

	return membership, nil
}

func (r *Repository) GetByUserAndWorkspaceID(
	ctx context.Context,
	userID uuid.UUID,
	workspaceID uuid.UUID,
) (*WorkspaceMembership, error) {
	query := `
		SELECT 
			id,
			workspace_id,
			user_id,
			role,
			created_at,
			updated_at
		FROM workspace_memberships
		WHERE user_id = $1 AND workspace_id = $2
	`

	membership := &WorkspaceMembership{}
	err := r.pool.QueryRow(
		ctx,
		query,
		userID,
		workspaceID,
	).Scan(
		&membership.ID,
		&membership.WorkspaceID,
		&membership.UserID,
		&membership.Role,
		&membership.CreatedAt,
		&membership.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get workspace membership by user and workspace ID: %w", err)
	}

	return membership, nil
}

func (r *Repository) GetMemberListByWorkspaceID(
	ctx context.Context,
	workspaceID uuid.UUID,
) ([]*WorkspaceMembership, error) {
	query := `
		SELECT 
			id,
			workspace_id,
			user_id,
			role,
			created_at,
			updated_at
		FROM workspace_memberships
		WHERE workspace_id = $1
	`

	rows, err := r.pool.Query(ctx, query, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace memberships by workspace ID: %w", err)
	}

	defer rows.Close()

	memberships := []*WorkspaceMembership{}

	for rows.Next() {
		membership := &WorkspaceMembership{}
		err := rows.Scan(
			&membership.ID,
			&membership.WorkspaceID,
			&membership.UserID,
			&membership.Role,
			&membership.CreatedAt,
			&membership.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workspace membership: %w", err)
		}
		memberships = append(memberships, membership)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred during rows iteration: %w", err)
	}

	return memberships, nil
}

func (r *Repository) ChangeMemberRole(
	ctx context.Context,
	membership *WorkspaceMembership,
) error {
	query := `
		UPDATE workspace_memberships
		SET role = $1, updated_at = NOW()
		WHERE id = $2 AND workspace_id = $3
	`

	_, err := r.pool.Exec(ctx, query, membership.Role, membership.ID, membership.WorkspaceID)
	if err != nil {
		return fmt.Errorf("failed to change member role: %w", err)
	}

	return nil
}

func (r *Repository) RemoveMember(
	ctx context.Context,
	workspaceMembership *WorkspaceMembership,
) error{
	query := `
		DELETE FROM workspace_memberships
		WHERE id = $1 AND workspace_id = $2
	`
	result, err := r.pool.Exec(ctx, query, workspaceMembership.ID, workspaceMembership.WorkspaceID)

	if err != nil {
		return fmt.Errorf("failed to remove member from workspace: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no member found with the given ID in the specified workspace")
	}
	return nil
}