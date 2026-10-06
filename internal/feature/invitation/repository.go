package invitation

import (
	"context"
	"fmt"
	"time"

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

func (r *Repository) CreateTx(
	ctx context.Context,
	tx pgx.Tx,
	invitation *WorkspaceInvitation,
) error {
	query := `
INSERT INTO workspace_invitations (
workspace_id,
invitee_user_id,
invited_by_user_id,
status,
expires_at
)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, created_at, updated_at
`

	return tx.QueryRow(
		ctx,
		query,
		invitation.WorkspaceID,
		invitation.InviteeUserID,
		invitation.InvitedByUserID,
		invitation.Status,
		invitation.ExpiresAt,
	).Scan(
		&invitation.ID,
		&invitation.CreatedAt,
		&invitation.UpdatedAt,
	)
}

func (r *Repository) GetByIDForUpdate(
	ctx context.Context,
	tx pgx.Tx,
	invitationID uuid.UUID,
) (*WorkspaceInvitation, error) {
	query := `
SELECT
id,
workspace_id,
invitee_user_id,
invited_by_user_id,
status,
created_at,
updated_at,
expires_at
FROM workspace_invitations
WHERE id = $1
FOR UPDATE
`

	invitation := &WorkspaceInvitation{}

	err := tx.QueryRow(
		ctx,
		query,
		invitationID,
	).Scan(
		&invitation.ID,
		&invitation.WorkspaceID,
		&invitation.InviteeUserID,
		&invitation.InvitedByUserID,
		&invitation.Status,
		&invitation.CreatedAt,
		&invitation.UpdatedAt,
		&invitation.ExpiresAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get workspace invitation by ID: %w", err)
	}

	return invitation, nil
}

func (r *Repository) UpdateStatusTx(
	ctx context.Context,
	tx pgx.Tx,
	invitationID uuid.UUID,
	status Status,
) (time.Time, error) {
	query := `
UPDATE workspace_invitations
SET status = $2, updated_at = NOW()
WHERE id = $1 AND status = 'pending'
RETURNING updated_at
`

	var updatedAt time.Time

	err := tx.QueryRow(
		ctx,
		query,
		invitationID,
		status,
	).Scan(&updatedAt)

	if err != nil {
		return time.Time{}, fmt.Errorf("failed to update workspace invitation status: %w", err)
	}

	return updatedAt, nil
}

func (r *Repository) ListByWorkspaceID(
	ctx context.Context,
	workspaceID uuid.UUID,
) ([]*InvitationResponse, error) {
	query := `
SELECT
i.id,
i.workspace_id,
i.invitee_user_id,
i.invited_by_user_id,
i.status,
u.email,
i.created_at,
i.updated_at,
i.expires_at
FROM workspace_invitations AS i
JOIN users AS u ON i.invitee_user_id = u.id
WHERE i.workspace_id = $1
ORDER BY i.created_at ASC
`

	rows, err := r.pool.Query(ctx, query, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace invitations by workspace ID: %w", err)
	}

	defer rows.Close()

	invitations := []*InvitationResponse{}

	for rows.Next() {
		invitation := &InvitationResponse{}
		if err := rows.Scan(
			&invitation.ID,
			&invitation.WorkspaceID,
			&invitation.InviteeUserID,
			&invitation.InvitedByUserID,
			&invitation.Status,
			&invitation.InviteeEmail,
			&invitation.CreatedAt,
			&invitation.UpdatedAt,
			&invitation.ExpiresAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan workspace invitation: %w", err)
		}

		invitations = append(invitations, invitation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred during invitation rows iteration: %w", err)
	}

	return invitations, nil
}
