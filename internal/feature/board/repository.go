package board

import (
	"context"

	"github.com/google/uuid"
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

func (r *Repository) Create(ctx context.Context, board *Board) error {
	query:=
	`INSERT INTO boards (
			name,
			workspace_id
		)
	VALUES ($1, $2)
	RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(ctx, query, board.Name, board.WorkspaceID).Scan(&board.ID, &board.CreatedAt, &board.UpdatedAt)
}

func (r *Repository) GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*BoardSummary, error) {
	query := 
	`SELECT
		id,
		name,
		workspace_id,
		created_at,
		updated_at
	FROM boards
	WHERE workspace_id = $1
	ORDER BY created_at ASC
	`	

	boards :=[]* BoardSummary{}
	rows, err := r.pool.Query(ctx, query, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query boards: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		board := &BoardSummary{}
		err := rows.Scan(
			&board.ID,
			&board.Name,
			&board.WorkspaceID,
			&board.CreatedAt,
			&board.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan board: %w", err)
		}
		boards = append(boards, board)
	}

	return boards, nil
}