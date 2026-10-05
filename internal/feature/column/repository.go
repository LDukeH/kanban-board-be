package column

import (
	"context"
	"errors"
	"fmt"

	"kanban-board-be/internal/apperrors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// db is the slice of *pgxpool.Pool this repository uses, so tests can inject a mock.
type db interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Repository struct {
	pool db
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

// NewRepositoryWithDB builds a repository over any db implementation, for tests.
func NewRepositoryWithDB(pool db) *Repository {
	return &Repository{pool: pool}
}

// classifyVersionMismatch splits "matched no rows" into conflict or not found.
func classifyVersionMismatch(exists bool) error {
	if exists {
		return apperrors.ErrConflict
	}
	return apperrors.ErrNotFound
}

// resolveVersionMismatch returns the error to propagate, so the caller rolls back.
func resolveVersionMismatch(ctx context.Context, q db, columnID uuid.UUID) error {
	var exists bool
	err := q.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM columns WHERE id = $1)`, columnID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check column existence: %w", err)
	}
	return classifyVersionMismatch(exists)
}

func (r *Repository) Create(ctx context.Context, column *Column) error {
	query :=
		`INSERT INTO columns (
		board_id,
		name,
		position
	)
	VALUES ($1, $2, COALESCE((SELECT MAX(position) + 1 FROM columns WHERE board_id = $1), 0))
	RETURNING id, position, version, created_at, updated_at`

	return r.pool.QueryRow(ctx, query, column.BoardID, column.Name).Scan(&column.ID, &column.Position, &column.Version, &column.CreatedAt, &column.UpdatedAt)
}

func (r *Repository) GetByBoardID(ctx context.Context, boardID uuid.UUID) ([]*Column, error) {
	query :=
		`SELECT
		id,
		board_id,
		name,
		position,
		version,
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
			&column.Version,
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

func (r *Repository) GetByID(ctx context.Context, columnID uuid.UUID) (*Column, error) {
	query :=
		`SELECT
		id,
		board_id,
		name,
		position,
		version,
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
		&column.Version,
		&column.CreatedAt,
		&column.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan column: %w", err)
	}
	return column, nil
}

// UpdateColumnName renames a column, guarded by the client's version.
func (r *Repository) UpdateColumnName(ctx context.Context, column *Column) error {
	query := `
		UPDATE columns
		SET name = $1,
		    version = version + 1,
		    updated_at = NOW()
		WHERE id = $2
		  AND version = $3
	`

	result, err := r.pool.Exec(ctx, query, column.Name, column.ID, column.Version)
	if err != nil {
		return fmt.Errorf("failed to update column: %w", err)
	}
	if result.RowsAffected() == 0 {
		return resolveVersionMismatch(ctx, r.pool, column.ID)
	}

	column.Version++
	return nil
}

// UpdateColumnPosition reorders a column, checking only the moved column's version.
func (r *Repository) UpdateColumnPosition(ctx context.Context, columnID uuid.UUID, newPosition int, version int) error {
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
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("failed to lock column: %w", err)
	}

	if newPosition == oldPosition {
		return nil
	}

	if newPosition > oldPosition {
		_, err = tx.Exec(ctx, `
            UPDATE columns
            SET position = position - 1,
                version = version + 1,
                updated_at = NOW()
            WHERE board_id = $1 AND position > $2 AND position <= $3
        `, boardID, oldPosition, newPosition)
	} else {
		_, err = tx.Exec(ctx, `
            UPDATE columns
            SET position = position + 1,
                version = version + 1,
                updated_at = NOW()
            WHERE board_id = $1 AND position >= $2 AND position < $3
        `, boardID, newPosition, oldPosition)
	}
	if err != nil {
		return fmt.Errorf("failed to reorder surrounding columns: %w", err)
	}

	result, err := tx.Exec(ctx, `
        UPDATE columns
        SET position = $1,
            version = version + 1,
            updated_at = NOW()
        WHERE id = $2
          AND version = $3
    `, newPosition, columnID, version)
	if err != nil {
		return fmt.Errorf("failed to update moved column: %w", err)
	}

	if result.RowsAffected() == 0 {
		// Returning here rolls the shift above back.
		return resolveVersionMismatch(ctx, tx, columnID)
	}

	return tx.Commit(ctx)
}

// DeleteColumn removes a column, guarded by the client's version.
func (r *Repository) DeleteColumn(ctx context.Context, columnID uuid.UUID, version int) error {
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
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("failed to get column before deletion: %w", err)
	}

	// Delete the column. Its tasks cascade through fk_tasks_column.
	result, err := tx.Exec(
		ctx,
		`
		DELETE FROM columns
		WHERE id = $1
		  AND version = $2
		`,
		columnID,
		version,
	)

	if err != nil {
		return fmt.Errorf("failed to delete column: %w", err)
	}

	if result.RowsAffected() == 0 {
		return resolveVersionMismatch(ctx, tx, columnID)
	}

	// Shift other columns' positions
	_, err = tx.Exec(
		ctx,
		`
		UPDATE columns
		SET position = position - 1,
		    version = version + 1,
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
