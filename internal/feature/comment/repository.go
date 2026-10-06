package comment

import (
	"context"
	"errors"
	"fmt"

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

func (r *Repository) ListByTaskID(ctx context.Context, taskID uuid.UUID) ([]*CommentResponse, error) {
	query := `
SELECT
c.id,
c.task_id,
c.author_id,
u.email AS author_email,
c.content,
c.created_at,
c.updated_at
FROM comments c
JOIN users u
ON u.id = c.author_id
WHERE c.task_id = $1
ORDER BY c.created_at ASC, c.id ASC
`

	rows, err := r.pool.Query(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments: %w", err)
	}
	defer rows.Close()

	comments := []*CommentResponse{}
	for rows.Next() {
		comment := &CommentResponse{}
		if err := rows.Scan(
			&comment.ID,
			&comment.TaskID,
			&comment.AuthorID,
			&comment.AuthorEmail,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over comment rows: %w", err)
	}

	return comments, nil
}

func (r *Repository) Create(ctx context.Context, comment *Comment) (*CommentResponse, error) {
	query := `
WITH inserted AS (
INSERT INTO comments (id, task_id, author_id, content)
VALUES ($1, $2, $3, $4)
RETURNING id, task_id, author_id, content, created_at, updated_at
)
SELECT
i.id,
i.task_id,
i.author_id,
u.email AS author_email,
i.content,
i.created_at,
i.updated_at
FROM inserted i
JOIN users u
ON u.id = i.author_id
`

	response := &CommentResponse{}
	err := r.pool.QueryRow(ctx, query, comment.ID, comment.TaskID, comment.AuthorID, comment.Content).Scan(
		&response.ID,
		&response.TaskID,
		&response.AuthorID,
		&response.AuthorEmail,
		&response.Content,
		&response.CreatedAt,
		&response.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	return response, nil
}

func (r *Repository) GetByID(ctx context.Context, commentID uuid.UUID) (*Comment, error) {
	query := `
SELECT
id,
task_id,
author_id,
content,
created_at,
updated_at
FROM comments
WHERE id = $1
`

	comment := &Comment{}
	err := r.pool.QueryRow(ctx, query, commentID).Scan(
		&comment.ID,
		&comment.TaskID,
		&comment.AuthorID,
		&comment.Content,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	return comment, nil
}

func (r *Repository) Delete(ctx context.Context, commentID uuid.UUID) error {
	commandTag, err := r.pool.Exec(ctx, `DELETE FROM comments WHERE id = $1`, commentID)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
