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

// GetBoardWorkspaceID returns the workspace ID for a given board.
// This satisfies the authz.BoardWorkspaceProvider interface.
func (r *Repository) GetBoardWorkspaceID(ctx context.Context, boardID uuid.UUID) (uuid.UUID, error) {
	query := `SELECT workspace_id FROM boards WHERE id = $1`

	var workspaceID uuid.UUID
	err := r.pool.QueryRow(ctx, query, boardID).Scan(&workspaceID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get board workspace ID: %w", err)
	}

	return workspaceID, nil
}

func (r *Repository) GetByID(ctx context.Context, boardID uuid.UUID) (*Board, error) {
	query := 
	`SELECT 
		id,
		name,
		workspace_id,
		created_at,
		updated_at
	FROM boards
	WHERE id = $1`

	row := r.pool.QueryRow(ctx, query, boardID)
	board := &Board{}
	err := row.Scan(
		&board.ID,
		&board.Name,
		&board.WorkspaceID,
		&board.CreatedAt,
		&board.UpdatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to scan board: %w", err)
	}
	return board, nil
}