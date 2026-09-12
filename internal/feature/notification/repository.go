package notification

import (
	"context"
	"encoding/json"
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

func (r *Repository) Create(ctx context.Context, notification *Notification) error {
	query := `
INSERT INTO notifications (
user_id,
type,
payload
)
VALUES ($1, $2, $3)
RETURNING id, created_at, updated_at
`

	return r.pool.QueryRow(ctx, query, notification.UserID, notification.Type, notification.Payload).Scan(
		&notification.ID,
		&notification.CreatedAt,
		&notification.UpdatedAt,
	)
}

func (r *Repository) CreateTx(
	ctx context.Context,
	tx pgx.Tx,
	notification *Notification,
) error {
	query := `
INSERT INTO notifications (
user_id,
type,
payload
)
VALUES ($1, $2, $3)
RETURNING id, created_at, updated_at
`

	return tx.QueryRow(ctx, query, notification.UserID, notification.Type, notification.Payload).Scan(
		&notification.ID,
		&notification.CreatedAt,
		&notification.UpdatedAt,
	)
}

func (r *Repository) ListByUserID(ctx context.Context, userID uuid.UUID, limit int, unreadOnly bool) ([]*Notification, error) {
	query := `
SELECT
id,
user_id,
type,
payload,
read_at,
created_at,
updated_at
FROM notifications
WHERE user_id = $1
`

	args := []any{userID}

	if unreadOnly {
		query += ` AND read_at IS NULL`
	}

	query += `
ORDER BY created_at DESC, id DESC
LIMIT $2
`

	args = append(args, limit)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer rows.Close()

	notifications := []*Notification{}

	for rows.Next() {
		notification := &Notification{}
		var payloadJSON []byte

		if err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&payloadJSON,
			&notification.ReadAt,
			&notification.CreatedAt,
			&notification.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		notification.Payload = json.RawMessage(payloadJSON)
		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over notification rows: %w", err)
	}

	return notifications, nil
}

func (r *Repository) CountUnreadByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
SELECT COUNT(*)
FROM notifications
WHERE user_id = $1
  AND read_at IS NULL
`

	var count int
	if err := r.pool.QueryRow(ctx, query, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	return count, nil
}

// Returns whether the notification belonged to the user (idempotent when already read).
func (r *Repository) MarkAsRead(ctx context.Context, notificationID uuid.UUID, userID uuid.UUID) (bool, error) {
	query := `
UPDATE notifications
SET read_at = COALESCE(read_at, NOW()),
updated_at = NOW()
WHERE id = $1
  AND user_id = $2
`

	ct, err := r.pool.Exec(ctx, query, notificationID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to mark notification as read: %w", err)
	}

	return ct.RowsAffected() > 0, nil
}

func (r *Repository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	query := `
UPDATE notifications
SET read_at = COALESCE(read_at, NOW()),
updated_at = NOW()
WHERE user_id = $1
  AND read_at IS NULL
`

	if _, err := r.pool.Exec(ctx, query, userID); err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}
