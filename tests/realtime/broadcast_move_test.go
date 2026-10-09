package realtime_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestBroadcastTaskMovedSameColumn(t *testing.T) {
	pool := testPool(t)
	f := newFixture(t, pool)

	taskID := uuid.New()
	inActorTx(t, pool, f.userID, func(tx pgx.Tx) {
		if _, err := tx.Exec(context.Background(),
			`INSERT INTO public.tasks (id, column_id, title, position) VALUES ($1, $2, 'moving', 0)`,
			taskID, f.sourceCol); err != nil {
			t.Fatalf("insert task: %v", err)
		}
	})
	clearCaptured(t, pool)

	inActorTx(t, pool, f.userID, func(tx pgx.Tx) {
		if _, err := tx.Exec(context.Background(),
			`UPDATE public.tasks SET position = 3, version = version + 1 WHERE id = $1`,
			taskID); err != nil {
			t.Fatalf("move task: %v", err)
		}
	})

	e := wantEnvelope(t, pool, f, "task_moved", taskID.String())
	if strField(t, e, "column_id") != f.sourceCol.String() {
		t.Fatalf("column_id = %v, want %s", e.Payload["column_id"], f.sourceCol)
	}
	if strField(t, e, "source_column_id") != f.sourceCol.String() {
		t.Fatalf("source_column_id = %v, want the same column %s", e.Payload["source_column_id"], f.sourceCol)
	}
}

func TestBroadcastTaskMovedCrossColumn(t *testing.T) {
	pool := testPool(t)
	f := newFixture(t, pool)

	taskID := uuid.New()
	inActorTx(t, pool, f.userID, func(tx pgx.Tx) {
		if _, err := tx.Exec(context.Background(),
			`INSERT INTO public.tasks (id, column_id, title, position) VALUES ($1, $2, 'cross', 0)`,
			taskID, f.sourceCol); err != nil {
			t.Fatalf("insert task: %v", err)
		}
	})
	clearCaptured(t, pool)

	inActorTx(t, pool, f.userID, func(tx pgx.Tx) {
		if _, err := tx.Exec(context.Background(),
			`UPDATE public.tasks SET column_id = $2, position = 0, version = version + 1 WHERE id = $1`,
			taskID, f.targetCol); err != nil {
			t.Fatalf("move task: %v", err)
		}
	})

	e := wantEnvelope(t, pool, f, "task_moved", taskID.String())
	if strField(t, e, "column_id") != f.targetCol.String() {
		t.Fatalf("column_id = %v, want %s", e.Payload["column_id"], f.targetCol)
	}
	if strField(t, e, "source_column_id") != f.sourceCol.String() {
		t.Fatalf("source_column_id = %v, want %s", e.Payload["source_column_id"], f.sourceCol)
	}
}

func TestBroadcastTaskDeleted(t *testing.T) {
	pool := testPool(t)
	f := newFixture(t, pool)

	taskID := uuid.New()
	inActorTx(t, pool, f.userID, func(tx pgx.Tx) {
		if _, err := tx.Exec(context.Background(),
			`INSERT INTO public.tasks (id, column_id, title, position) VALUES ($1, $2, 'doomed', 0)`,
			taskID, f.sourceCol); err != nil {
			t.Fatalf("insert task: %v", err)
		}
	})
	clearCaptured(t, pool)

	inActorTx(t, pool, f.userID, func(tx pgx.Tx) {
		if _, err := tx.Exec(context.Background(),
			`DELETE FROM public.tasks WHERE id = $1`, taskID); err != nil {
			t.Fatalf("delete task: %v", err)
		}
	})

	wantEnvelope(t, pool, f, "task_deleted", taskID.String())
}
