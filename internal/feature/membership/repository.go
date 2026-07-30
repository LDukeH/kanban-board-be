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