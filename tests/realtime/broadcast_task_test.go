package realtime_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestBroadcastTaskCreated(t *testing.T) {
	pool := testPool(t)
	f := newFixture(t, pool)
	clearCaptured(t, pool)

	taskID := uuid.New()
	inActorTx(t, pool, f.userID, func(tx pgx.Tx) {
		if _, err := tx.Exec(context.Background(),
			`INSERT INTO public.tasks (id, column_id, title, position) VALUES ($1, $2, 'fresh', 0)`,
			taskID, f.sourceCol); err != nil {
			t.Fatalf("insert task: %v", err)
		}
	})

	e := wantEnvelope(t, pool, f, "task_created", taskID.String())
	if strField(t, e, "column_id") != f.sourceCol.String() {
		t.Fatalf("column_id = %v, want %s", e.Payload["column_id"], f.sourceCol)
	}
	if e.Payload["version"].(float64) != 1 {
		t.Fatalf("version = %v, want 1", e.Payload["version"])
	}
	if v, ok := e.Payload["source_column_id"]; ok && v != nil {
		t.Fatalf("source_column_id = %v, want null on create", v)
	}
}

func TestBroadcastTaskUpdated(t *testing.T) {
	pool := testPool(t)
	f := newFixture(t, pool)

	taskID := uuid.New()
	inActorTx(t, pool, f.userID, func(tx pgx.Tx) {
		if _, err := tx.Exec(context.Background(),
			`INSERT INTO public.tasks (id, column_id, title, position) VALUES ($1, $2, 'before', 0)`,
			taskID, f.sourceCol); err != nil {
			t.Fatalf("insert task: %v", err)
		}
	})
	clearCaptured(t, pool)

	inActorTx(t, pool, f.userID, func(tx pgx.Tx) {
		if _, err := tx.Exec(context.Background(),
			`UPDATE public.tasks SET title = 'after', version = version + 1 WHERE id = $1`,
			taskID); err != nil {
			t.Fatalf("update task: %v", err)
		}
	})

	e := wantEnvelope(t, pool, f, "task_updated", taskID.String())
	if e.Payload["version"].(float64) != 2 {
		t.Fatalf("version = %v, want 2", e.Payload["version"])
	}
}
