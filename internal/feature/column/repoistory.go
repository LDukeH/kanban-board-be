package column

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

func (r *Repository) Create(ctx context.Context, column *Column) error {
	query := 
	`INSERT INTO columns (
		board_id,
		name,
		position
	)
	VALUES ($1, $2, $3)
	RETURNING id, created_at, updated_at`
	
	return r.pool.QueryRow(ctx, query, column.BoardID, column.Name, column.Position).Scan(&column.ID, &column.CreatedAt, &column.UpdatedAt)
}

func (r *Repository) GetByBoardID(ctx context.Context, boardID uuid.UUID) ([]*Column, error) {	
	query:=
	`SELECT
		id,
		board_id,
		name,
		position,
		created_at,
		updated_at
	FROM columns
	WHERE board_id = $1
	ORDER BY position ASC
	`

	columns := []*Column{}
	rows, err := r.pool.Query(ctx, query, boardID)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		column := &Column{}
		err := rows.Scan(
			&column.ID,
			&column.BoardID,
			&column.Name,
			&column.Position,
			&column.CreatedAt,
			&column.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}

	return columns, nil
}

// GetColumnBoardID returns the board ID for a given column.
// This satisfies the authz.ColumnBoardProvider interface.
func (r *Repository) GetColumnBoardID(ctx context.Context, columnID uuid.UUID) (uuid.UUID, error) {
	query := `SELECT board_id FROM columns WHERE id = $1`

	var boardID uuid.UUID
	err := r.pool.QueryRow(ctx, query, columnID).Scan(&boardID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get column board ID: %w", err)
	}

	return boardID, nil
}

func (r *Repository) GetByID(ctx context.Context, columnID string) (*Column, error) {
	query := 
	`SELECT
		id,
		board_id,
		name,
		position,
		created_at,
		updated_at
	FROM columns
	WHERE id = $1`

	row := r.pool.QueryRow(ctx, query, columnID)
	column := &Column{}
	err := row.Scan(
		&column.ID,
		&column.BoardID,
		&column.Name,
		&column.Position,
		&column.CreatedAt,
		&column.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan column: %w", err)
	}
	return column, nil
}