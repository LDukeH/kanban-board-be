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
	VALUES ($1, $2, COALESCE((SELECT MAX(position) + 1 FROM columns WHERE board_id = $1), 0))
	RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(ctx, query, column.BoardID, column.Name).Scan(&column.ID, &column.CreatedAt, &column.UpdatedAt)
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
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over column rows: %w", err)
	}

	return columns, nil
}

// GetColumnBoardID returns the board ID for a given column.
// This satisfies the authz.ColumnBoardProvider interface.
func (r *Repository) GetColumnBoardID(ctx context.Context, columnID uuid.UUID) (uuid.UUID, error) {
	query := `SELECT board_id FROM columns WHERE id = $1`
	fmt.Printf("GetColumnBoardID called with columnID: %s", columnID)
	var boardID uuid.UUID
	err := r.pool.QueryRow(ctx, query, columnID).Scan(&boardID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get column board ID: %w", err)
	}

	return boardID, nil
}

func (r *Repository) GetByID(ctx context.Context, columnID uuid.UUID) (*Column, error) {
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

func (r *Repository) UpdateColumnName(ctx context.Context, column *Column) error {
	query := `
		UPDATE columns
		SET name = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.pool.Exec(ctx, query, column.Name, column.ID)
	if err != nil {
		return fmt.Errorf("failed to update column: %w", err)
	}
	return nil
}

func (r *Repository) UpdateColumnPosition(ctx context.Context, columnID uuid.UUID, newPosition int) error {
    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback(ctx)

    var boardID uuid.UUID
    var oldPosition int

    // Lock row inside transaction
    err = tx.QueryRow(
        ctx,
        `SELECT board_id, position FROM columns WHERE id = $1 FOR UPDATE`,
        columnID,
    ).Scan(&boardID, &oldPosition)
    if err != nil {
        return fmt.Errorf("failed to lock column: %w", err)
    }

    if newPosition == oldPosition {
        return nil
    }

    if newPosition > oldPosition {
        _, err = tx.Exec(ctx, `
            UPDATE columns
            SET position = position - 1, updated_at = NOW()
            WHERE board_id = $1 AND position > $2 AND position <= $3
        `, boardID, oldPosition, newPosition)
    } else {
        _, err = tx.Exec(ctx, `
            UPDATE columns
            SET position = position + 1, updated_at = NOW()
            WHERE board_id = $1 AND position >= $2 AND position < $3
        `, boardID, newPosition, oldPosition)
    }
    if err != nil {
        return fmt.Errorf("failed to reorder surrounding columns: %w", err)
    }

    _, err = tx.Exec(ctx, `
        UPDATE columns SET position = $1, updated_at = NOW() WHERE id = $2
    `, newPosition, columnID)
    if err != nil {
        return fmt.Errorf("failed to update moved column: %w", err)
    }

    return tx.Commit(ctx)
}

func (r *Repository) DeleteColumn(ctx context.Context, columnID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get the position and board id
	var boardID uuid.UUID
	var position int

	err = tx.QueryRow(
		ctx,
		`
		SELECT board_id, position
		FROM columns
		WHERE id = $1
		`,
		columnID,
	).Scan(&boardID, &position)

	if err != nil {
		return fmt.Errorf("failed to get column before deletion: %w", err)
	}

	// Delete the column.
	_, err = tx.Exec(
		ctx,
		`
		DELETE FROM columns
		WHERE id = $1
		`,
		columnID,
	)

	if err != nil {
		return fmt.Errorf("failed to delete column: %w", err)
	}

	// Shift other columns' positions
	_, err = tx.Exec(
		ctx,
		`
		UPDATE columns
		SET position = position - 1,
		    updated_at = NOW()
		WHERE board_id = $1
		  AND position > $2
		`,
		boardID,
		position,
	)

	if err != nil {
		return fmt.Errorf("failed to reorder columns after deletion: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit column deletion: %w", err)
	}

	return nil
}