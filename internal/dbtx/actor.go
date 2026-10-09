package dbtx

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)
type Execer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}


func SetActor(ctx context.Context, e Execer, userID uuid.UUID) error {
	_, err := e.Exec(ctx, `SELECT set_config('app.actor_id', $1, true)`, userID.String())
	return err
}
