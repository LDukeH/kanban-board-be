package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

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

// taskExists reports whether the task row is still present.
func taskExists(ctx context.Context, q db, taskID uuid.UUID) (bool, error) {
	var exists bool
	err := q.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM tasks WHERE id = $1)`, taskID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check task existence: %w", err)
	}
	return exists, nil
}

// resolveVersionMismatch returns the error to propagate, so the caller rolls back.
func resolveVersionMismatch(ctx context.Context, q db, taskID uuid.UUID) error {
	exists, err := taskExists(ctx, q, taskID)
	if err != nil {
		return err
	}
	return classifyVersionMismatch(exists)
}

func (r *Repository) Create(ctx context.Context, task *Task) error {
	query :=
		`INSERT INTO tasks (
			column_id,
			title,
			description,
			position,
			priority,
			due_date,
			completed
		)
		VALUES ($1, $2, $3, COALESCE((SELECT MAX(position) + 1 FROM tasks WHERE column_id = $1), 0), COALESCE($4, 'none'), $5, COALESCE($6, false))
		RETURNING id, priority, due_date, completed, version, created_at, updated_at`

	var dueDate any
	if d, ok := task.DueDate.Get(); ok {
		dueDate = d.Time
	}

	var priority string
	var completed bool
	var dueDateSQL *time.Time
	err := r.pool.QueryRow(ctx, query, task.ColumnID, task.Title, task.Description, task.Priority, dueDate, task.Completed).Scan(
		&task.ID,
		&priority,
		&dueDateSQL,
		&completed,
		&task.Version,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		return err
	}
	task.Priority = &priority
	task.Completed = &completed
	task.DueDate = dateToOptional(dueDateSQL)
	return nil
}

func dateToOptional(t *time.Time) Optional[DateOnly] {
	if t == nil {
		return Null[DateOnly]()
	}
	return Some(DateOnly{Time: *t})
}

func optionalDateArg(o Optional[DateOnly]) any {
	if d, ok := o.Get(); ok {
		return d.Time
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, taskID uuid.UUID) (*Task, error) {
	query :=
		`SELECT
			id,
			column_id,
			title,
			description,
			position,
			priority,
			due_date,
			completed,
			version,
			created_at,
			updated_at
		FROM tasks
		WHERE id = $1`

	row := r.pool.QueryRow(ctx, query, taskID)
	task := &Task{}
	var dueDateSQL *time.Time
	err := row.Scan(
		&task.ID,
		&task.ColumnID,
		&task.Title,
		&task.Description,
		&task.Position,
		&task.Priority,
		&dueDateSQL,
		&task.Completed,
		&task.Version,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan task: %w", err)
	}
	task.DueDate = dateToOptional(dueDateSQL)
	return task, nil
}

func (r *Repository) GetByColumnID(ctx context.Context, columnID uuid.UUID) ([]*Task, error) {
	query := `
		WITH column_tasks AS (
			SELECT id
			FROM tasks
			WHERE column_id = $1
		),
		task_labels_agg AS (
			SELECT
				tl.task_id,
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
				) AS labels
			FROM task_labels tl
			JOIN labels l ON l.id = tl.label_id
			WHERE tl.task_id IN (SELECT id FROM column_tasks)
			GROUP BY tl.task_id
		),
		task_assignees_agg AS (
			SELECT
				ta.task_id,
				json_agg(
					json_build_object(
						'user_id', u.id,
						'email', u.email
					)
					ORDER BY u.email
				) AS assignees
			FROM task_assignees ta
			JOIN users u ON u.id = ta.user_id
			WHERE ta.task_id IN (SELECT id FROM column_tasks)
			GROUP BY ta.task_id
		)
		SELECT
			t.id,
			t.column_id,
			t.title,
			t.description,
			t.position,
			t.priority,
			t.due_date,
			t.completed,
			t.version,
			t.created_at,
			t.updated_at,
			COALESCE(tla.labels, '[]'::json) AS labels,
			COALESCE(taa.assignees, '[]'::json) AS assignees
		FROM tasks t
		LEFT JOIN task_labels_agg tla ON tla.task_id = t.id
		LEFT JOIN task_assignees_agg taa ON taa.task_id = t.id
		WHERE t.column_id = $1
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
		var assigneesJSON []byte
		var dueDateSQL *time.Time

		err := rows.Scan(
			&task.ID,
			&task.ColumnID,
			&task.Title,
			&task.Description,
			&task.Position,
			&task.Priority,
			&dueDateSQL,
			&task.Completed,
			&task.Version,
			&task.CreatedAt,
			&task.UpdatedAt,
			&labelsJSON,
			&assigneesJSON,
		)
		if err != nil {
			return nil, err
		}
		task.DueDate = dateToOptional(dueDateSQL)

		if err := json.Unmarshal(labelsJSON, &task.Labels); err != nil {
			return nil, fmt.Errorf("error unmarshalling task labels: %w", err)
		}

		if err := json.Unmarshal(assigneesJSON, &task.Assignees); err != nil {
			return nil, fmt.Errorf("error unmarshalling task assignees: %w", err)
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over tasks rows: %w", err)
	}

	return tasks, nil
}

// UpdateTask applies a partial update guarded by the client's version.
func (r *Repository) UpdateTask(ctx context.Context, task *Task) error {
	setClauses, args := buildTaskUpdateSet(task)
	if len(setClauses) == 0 {
		// An empty patch still has to honour the guard.
		return r.guardVersion(ctx, task)
	}

	args = append(args, task.ID)      // The task ID
	args = append(args, task.Version) // The version the client read

	query := fmt.Sprintf(`
		UPDATE tasks
		SET %s,
			version = version + 1,
			updated_at = NOW()
		WHERE id = $%d
		  AND version = $%d
	`,
		strings.Join(setClauses, ", "),
		len(args)-1, // The position of the task ID in the args slice
		len(args),   // The position of the version in the args slice
	)

	result, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	if result.RowsAffected() == 0 {
		return resolveVersionMismatch(ctx, r.pool, task.ID)
	}

	task.Version++
	return nil
}

func (r *Repository) guardVersion(ctx context.Context, task *Task) error {
	var current int
	err := r.pool.QueryRow(ctx, `SELECT version FROM tasks WHERE id = $1`, task.ID).Scan(&current)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("failed to read task version: %w", err)
	}
	if current != task.Version {
		return apperrors.ErrConflict
	}
	return nil
}

func buildTaskUpdateSet(task *Task) ([]string, []any) {
	var setClauses []string
	var args []any

	addArg := func(clause string, value any) {
		setClauses = append(setClauses, fmt.Sprintf(clause, len(args)+1))
		args = append(args, value)
	}

	if task.Title != nil {
		addArg("title = $%d", *task.Title)
	}

	if task.Description != nil {
		addArg("description = $%d", *task.Description)
	}

	if task.Priority != nil {
		addArg("priority = $%d", *task.Priority)
	}

	if task.Completed != nil {
		addArg("completed = $%d", *task.Completed)
	}

	switch task.DueDate.State() {
	case OptionalNull:
		setClauses = append(setClauses, "due_date = NULL")
	case OptionalValue:
		if d, ok := task.DueDate.Get(); ok {
			addArg("due_date = $%d", d.Time)
		}
	case OptionalUnset:
		// Absent key: leave the stored value unchanged.
	}

	return setClauses, args
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
		SELECT id, column_id, position, version
		FROM tasks
		WHERE id = $1
		FOR UPDATE
		`,
		task.ID,
	).Scan(
		&currentTask.ID,
		&currentTask.ColumnID,
		&currentTask.Position,
		&currentTask.Version,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("failed to get current task: %w", err)
	}

	sourceColumnID := currentTask.ColumnID
	targetColumnID := task.ColumnID

	oldPosition := currentTask.Position
	newPosition := task.Position

	// No change in position or column
	if newPosition == oldPosition && targetColumnID == sourceColumnID {
		task.Version = currentTask.Version
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
					    version = version + 1,
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
					    version = version + 1,
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
				    version = version + 1,
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
				    version = version + 1,
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

	result, err := tx.Exec(
		ctx,
		`
			UPDATE tasks
			SET position = $1,
				column_id = $2,
				version = version + 1,
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

	if result.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit task move: %w", err)
	}

	task.Version = currentTask.Version + 1
	return nil
}

// DeleteTask removes a task guarded by the client's version and closes the gap.
func (r *Repository) DeleteTask(ctx context.Context, taskID uuid.UUID, version int) error {
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
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("failed to get task before deletion: %w", err)
	}

	// Delete the task, guarded by the version the client read
	result, err := tx.Exec(
		ctx,
		`
		DELETE FROM tasks
		WHERE id = $1
		  AND version = $2
		`,
		taskID,
		version,
	)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	if result.RowsAffected() == 0 {
		return resolveVersionMismatch(ctx, tx, taskID)
	}

	// Shift other tasks' positions
	_, err = tx.Exec(
		ctx,
		`
		UPDATE tasks
		SET position = position - 1,
		    version = version + 1,
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

func (r *Repository) AssignUserToTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error {
	query := `
		INSERT INTO task_assignees (task_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`

	_, err := r.pool.Exec(ctx, query, taskID, userID)
	if err != nil {
		return fmt.Errorf("failed to assign user to task: %w", err)
	}

	return nil
}

func (r *Repository) RemoveUserFromTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error {
	query := `
		DELETE FROM task_assignees
		WHERE task_id = $1 AND user_id = $2
	`

	_, err := r.pool.Exec(ctx, query, taskID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove user from task: %w", err)
	}

	return nil
}

func (r *Repository) IsUserAssignedToTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM task_assignees
			WHERE task_id = $1 AND user_id = $2
		)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, taskID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user assignment: %w", err)
	}

	return exists, nil
}

func (r *Repository) GetUsersByTaskID(ctx context.Context, taskID uuid.UUID) ([]*UsersTaskResponse, error) {
	query := `
		SELECT u.id, u.email
		FROM users u
		JOIN task_assignees ta ON u.id = ta.user_id
		WHERE ta.task_id = $1
	`
	rows, err := r.pool.Query(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by task ID: %w", err)
	}
	defer rows.Close()

	users := []*UsersTaskResponse{}
	for rows.Next() {
		user := &UsersTaskResponse{}
		err := rows.Scan(&user.UserID, &user.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over user rows: %w", err)
	}
	return users, nil
}

func (r *Repository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

func (r *Repository) AssignUserToTaskTx(ctx context.Context, tx pgx.Tx, taskID uuid.UUID, userID uuid.UUID) error {
	query := `
INSERT INTO task_assignees (task_id, user_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING
`

	_, err := tx.Exec(ctx, query, taskID, userID)
	if err != nil {
		return fmt.Errorf("failed to assign user to task: %w", err)
	}

	return nil
}
