package task

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

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

func (r *Repository) Create(ctx context.Context, task *Task) error {
	query :=
		`INSERT INTO tasks (
			column_id,
			title,
			description,
			position
		)
		VALUES ($1, $2, $3, COALESCE((SELECT MAX(position) + 1 FROM tasks WHERE column_id = $1), 0))
		RETURNING id, created_at, updated_at`

		return r.pool.QueryRow(ctx, query, task.ColumnID, task.Title, task.Description).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
}

func (r *Repository) GetByID(ctx context.Context, taskID uuid.UUID) (*Task, error) {
	query :=
		`SELECT
			id,
			column_id,
			title,
			description,
			position,
			created_at,
			updated_at
		FROM tasks
		WHERE id = $1`

	row := r.pool.QueryRow(ctx, query, taskID)
	task := &Task{}
	err := row.Scan(
		&task.ID,
		&task.ColumnID,
		&task.Title,
		&task.Description,
		&task.Position,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan task: %w", err)
	}
	return task, nil
}

func (r *Repository) GetByColumnID(ctx context.Context, columnID uuid.UUID) ([]*Task, error) {
	query := `
	SELECT
		t.id,
		t.column_id,
		t.title,
		t.description,
		t.position,
		t.created_at,
		t.updated_at,
		COALESCE(
			json_agg(
				json_build_object(
					'id', l.id,
					'board_id', l.board_id,
					'name', l.name,
					'color', l.color,
					'created_at', l.created_at,
					'updated_at', l.updated_at
				)
				ORDER BY l.name
			) FILTER (WHERE l.id IS NOT NULL),
			'[]'
		) AS labels
	FROM tasks t
	LEFT JOIN task_labels tl
		ON tl.task_id = t.id
	LEFT JOIN labels l
		ON l.id = tl.label_id
	WHERE t.column_id = $1
	GROUP BY
		t.id,
		t.column_id,
		t.title,
		t.description,
		t.position,
		t.created_at,
		t.updated_at
	ORDER BY t.position ASC
	`
	tasks := []*Task{}
	rows, err := r.pool.Query(ctx, query, columnID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		task := &Task{}
		var labelsJSON []byte

		err := rows.Scan(
			&task.ID,
			&task.ColumnID,
			&task.Title,
			&task.Description,
			&task.Position,
			&task.CreatedAt,
			&task.UpdatedAt,
			&labelsJSON,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(labelsJSON, &task.Labels); err != nil {
			return nil, fmt.Errorf("error unmarshalling task labels: %w", err)
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over tasks rows: %w", err)
	}

	return tasks, nil
}

func (r *Repository) UpdateTask(ctx context.Context, task *Task) error {
	var setClauses []string
	var args []interface{}
	
	if( task.Title != nil ) {
		setClauses = append(setClauses, "title = $1")
		args = append(args, *task.Title)
	}

	if( task.Description != nil ) {
		setClauses = append(setClauses, "description = $2")
		args = append(args, *task.Description)
	}

	if(len(setClauses) == 0) {
		return nil // Nothing to update
	}

	args = append(args, task.ID) // Add the task ID as the last argument

	query := fmt.Sprintf(`
		UPDATE tasks
		SET %s,
			updated_at = NOW()
		WHERE id = $%d
	`,
		strings.Join(setClauses, ", "),
		len(args), // The position of the task ID in the args slice
	)
	
	result, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no task found with ID %s", task.ID)
	}

	return nil
}

func (r *Repository) UpdateTaskPosition(ctx context.Context, task *Task) error {
	tx, err := r.pool.Begin(ctx)

	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var currentTask Task

	err = tx.QueryRow(
		ctx,
		`
		SELECT id, column_id, position
		FROM tasks
		WHERE id = $1
		`,
		task.ID,
	).Scan(
		&currentTask.ID,
		&currentTask.ColumnID,
		&currentTask.Position,
	)
	if err != nil {
		return fmt.Errorf("failed to get current task: %w", err)
	}

	sourceColumnID := currentTask.ColumnID
	targetColumnID := task.ColumnID

	oldPosition := currentTask.Position
	newPosition := task.Position

	// No change in position or column
	if newPosition == oldPosition && targetColumnID == sourceColumnID {
		return nil
	}

	if targetColumnID == sourceColumnID {
		// Reorder within the same column
		if newPosition > oldPosition {
			_, err = tx.Exec(
				ctx,
				`
					UPDATE tasks
					SET position = position - 1,
						updated_at = NOW()
					WHERE column_id = $1
					  AND position > $2
					  AND position <= $3
				`,
				sourceColumnID,
				oldPosition,
				newPosition,
			)
		} else {
			_, err = tx.Exec(
				ctx,
				`
					UPDATE tasks
					SET position = position + 1,
						updated_at = NOW()
					WHERE column_id = $1
					  AND position >= $2
					  AND position < $3
				`,
				sourceColumnID,
				newPosition,
				oldPosition,
			)
		}
		if err != nil {
			return fmt.Errorf("failed to reorder tasks: %w", err)
		}
	} else {
		// Cross-column movement: close gap in source column
		_, err = tx.Exec(
			ctx,
			`
				UPDATE tasks
				SET position = position - 1,
					updated_at = NOW()
				WHERE column_id = $1
				  AND position > $2
			`,
			sourceColumnID,
			oldPosition,
		)
		if err != nil {
			return fmt.Errorf("failed to close gap in source column: %w", err)
		}

		// Make space in target column
		_, err = tx.Exec(
			ctx,
			`
				UPDATE tasks
				SET position = position + 1,
					updated_at = NOW()
				WHERE column_id = $1
				  AND position >= $2
			`,
			targetColumnID,
			newPosition,
		)
		if err != nil {
			return fmt.Errorf("failed to make space in target column: %w", err)
		}
	}

	// Move target task (update position and optionally column_id)
	_, err = tx.Exec(
		ctx,
		`
			UPDATE tasks
			SET position = $1,
				column_id = $2,
				updated_at = NOW()
			WHERE id = $3
		`,
		newPosition,
		targetColumnID,
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update task position: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit task move: %w", err)
	}

	return nil
}

// GetTaskColumnID returns the column ID for a given task.
// This satisfies the authz.TaskColumnProvider interface.
func (r *Repository) GetTaskColumnID(ctx context.Context, taskID uuid.UUID) (uuid.UUID, error) {
	query := `SELECT column_id FROM tasks WHERE id = $1`

	var columnID uuid.UUID
	err := r.pool.QueryRow(ctx, query, taskID).Scan(&columnID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get task column ID: %w", err)
	}

	return columnID, nil
}

func (r *Repository) DeleteTask(ctx context.Context, taskID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get the position and column id of the task to be deleted
	var columnID uuid.UUID
	var position int

	err = tx.QueryRow(
		ctx,
		`
		SELECT column_id, position
		FROM tasks
		WHERE id = $1
		`,
		taskID,
	).Scan(&columnID, &position)

	if err != nil {
		return fmt.Errorf("failed to get task before deletion: %w", err)
	}

	// Delete the task
	_, err = tx.Exec(
		ctx,
		`
		DELETE FROM tasks
		WHERE id = $1
		`,
		taskID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	// Shift other tasks' positions
	_, err = tx.Exec(
		ctx,
		`
		UPDATE tasks
		SET position = position - 1,
		    updated_at = NOW()
		WHERE column_id = $1
		  AND position > $2
		`,
		columnID,
		position,
	)
	if err != nil {
		return fmt.Errorf("failed to reorder tasks after deletion: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit task deletion: %w", err)
	}

	return nil
}