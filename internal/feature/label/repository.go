package label

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *repository {
	return &repository{
		pool: pool,
	}
}

func (r *repository) Create(ctx context.Context, label *Label) error {
	query := `
	INSERT INTO labels (
		board_id,
		name,
		color
	)
	VALUES ($1, $2, $3)
	RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(ctx, query, label.BoardID, label.Name, label.Color).Scan(&label.ID, &label.CreatedAt, &label.UpdatedAt)
}

func (r *repository) GetLabelsByBoardID(ctx context.Context, boardID uuid.UUID) ([]*Label, error) {
	query := `
	SELECT 
		id,
		board_id,
		name,
		color,
		created_at,
		updated_at
	FROM labels
	WHERE board_id = $1
	`

	labels := []*Label{}
	rows, err := r.pool.Query(ctx, query, boardID)
	if err != nil {
		return nil, fmt.Errorf("failed to query labels: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		label := &Label{}
		err := rows.Scan(
			&label.ID,
			&label.BoardID,
			&label.Name,
			&label.Color,
			&label.CreatedAt,
			&label.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan label: %w", err)
		}
		labels = append(labels, label)
	}

	return labels, nil
}

func (r *repository) GetLabelByID(ctx context.Context, labelID uuid.UUID) (*Label, error) {
	query := `
	SELECT 
		id,
		board_id,
		name,
		color,
		created_at,
		updated_at
	FROM labels
	WHERE id = $1
	`
	label := &Label{}
	err := r.pool.QueryRow(ctx, query, labelID).Scan(
		&label.ID,
		&label.BoardID,
		&label.Name,
		&label.Color,
		&label.CreatedAt,
		&label.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query label by ID: %w", err)
	}
	return label, nil
}

func (r *repository) UpdateLabel(ctx context.Context, label *Label) error {
	query := `
	UPDATE labels
	SET name = $1, color = $2, updated_at = NOW()
	WHERE id = $3
	RETURNING updated_at`
	return r.pool.QueryRow(ctx, query, label.Name, label.Color, label.ID).Scan(&label.UpdatedAt)
}

func (r *repository) DeleteLabel(ctx context.Context, labelID uuid.UUID) error {
	query := `
	DELETE FROM labels
	WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, labelID)

	return err
}

// GetLabelBoardID returns the board ID for a given label.
// This satisfies the authz.LabelBoardProvider interface.
func (r *repository) GetLabelBoardID(ctx context.Context, labelID uuid.UUID) (uuid.UUID, error) {
	query := `SELECT board_id FROM labels WHERE id = $1`
	var boardID uuid.UUID
	err := r.pool.QueryRow(ctx, query, labelID).Scan(&boardID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get label board ID: %w", err)
	}
	return boardID, nil
}

// This is to assign a label to a task. It creates a new entry in the task_labels table.
func (r *repository) AddLabelToTask(ctx context.Context, taskID, labelID uuid.UUID) error {
	query := `
	INSERT INTO task_labels (task_id, label_id)
	VALUES ($1, $2)
	RETURNING task_id, label_id`

	_, err := r.pool.Exec(ctx, query, taskID, labelID)
	return err
}

func (r *repository) RemoveLabelFromTask(ctx context.Context, taskID, labelID uuid.UUID) error {
	query := `
	DELETE FROM task_labels
	WHERE task_id = $1 AND label_id = $2`

	_, err := r.pool.Exec(ctx, query, taskID, labelID)
	return err
}

func (r *repository) GetLabelsByTaskID(ctx context.Context, taskID uuid.UUID) ([]*Label, error) {
	query := `
	SELECT
		l.id,
		l.board_id,
		l.name,
		l.color,
		l.created_at,
		l.updated_at
	FROM task_labels tl
	JOIN labels l ON l.id = tl.label_id
	WHERE tl.task_id = $1
	`

	labels := []*Label{}
	rows, err := r.pool.Query(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query labels for task: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		label := &Label{}
		err := rows.Scan(
			&label.ID,
			&label.BoardID,
			&label.Name,
			&label.Color,
			&label.CreatedAt,
			&label.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan label: %w", err)
		}
		labels = append(labels, label)
	}

	return labels, nil
}
