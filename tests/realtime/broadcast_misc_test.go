package realtime_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestBroadcastReorderEmitsShiftedNeighbours(t *testing.T) {
	pool := testPool(t)
	f := newFixture(t, pool)

	a, b, c := uuid.New(), uuid.New(), uuid.New()
	inActorTx(t, pool, f.userID, func(tx pgx.Tx) {
		ctx := context.Background()
		for i, id := range []uuid.UUID{a, b, c} {
			if _, err := tx.Exec(ctx,
				`INSERT INTO public.tasks (id, column_id, title, position) VALUES ($1, $2, $3, $4)`,
				id, f.sourceCol, "t", i); err != nil {
				t.Fatalf("insert task: %v", err)
			}
		}
	})
	clearCaptured(t, pool)

	// Drag c from position 2 to position 0 and shift its neighbours down.
	inActorTx(t, pool, f.userID, func(tx pgx.Tx) {
		ctx := context.Background()
		if _, err := tx.Exec(ctx,
			`UPDATE public.tasks SET position = 0, version = version + 1 WHERE id = $1`, c); err != nil {
			t.Fatalf("move dragged task: %v", err)
		}
		if _, err := tx.Exec(ctx,
			`UPDATE public.tasks SET position = position + 1, version = version + 1
			 WHERE column_id = $1 AND id <> $2 AND position < 2`, f.sourceCol, c); err != nil {
			t.Fatalf("shift neighbours: %v", err)
		}
	})

	events := capturedEvents(t, pool)
	if len(events) != 3 {
		t.Fatalf("captured %d events, want 3 (dragged task plus two shifted neighbours): %+v", len(events), events)
	}
	seen := map[string]bool{}
	for _, e := range events {
		if e.Topic != "board:"+f.boardID.String() || e.Type != "task_moved" {
			t.Fatalf("unexpected reorder event: %+v", e)
		}
		if strField(t, e, "actor_id") != f.userID.String() {
			t.Fatalf("actor_id = %v, want %s", e.Payload["actor_id"], f.userID)
		}
		seen[strField(t, e, "entity_id")] = true
	}
	for _, id := range []uuid.UUID{a, b, c} {
		if !seen[id.String()] {
			t.Fatalf("missing reorder event for %s: %+v", id, events)
		}
	}
}

func TestBroadcastColumnDeletedSkipsCascadedTasks(t *testing.T) {
	pool := testPool(t)
	f := newFixture(t, pool)

	columnID := uuid.New()
	inActorTx(t, pool, f.userID, func(tx pgx.Tx) {
		ctx := context.Background()
		if _, err := tx.Exec(ctx,
			`INSERT INTO public.columns (id, board_id, name, position) VALUES ($1, $2, 'gone', 2)`,
			columnID, f.boardID); err != nil {
			t.Fatalf("insert column: %v", err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO public.tasks (id, column_id, title, position) VALUES ($1, $2, 'orphan', 0)`,
			uuid.New(), columnID); err != nil {
			t.Fatalf("insert task: %v", err)
		}
	})
	clearCaptured(t, pool)

	inActorTx(t, pool, f.userID, func(tx pgx.Tx) {
		if _, err := tx.Exec(context.Background(),
			`DELETE FROM public.columns WHERE id = $1`, columnID); err != nil {
			t.Fatalf("delete column: %v", err)
		}
	})

	e := wantEnvelope(t, pool, f, "column_deleted", columnID.String())
	if strField(t, e, "column_id") != columnID.String() {
		t.Fatalf("column_id = %v, want %s", e.Payload["column_id"], columnID)
	}
}

func TestBroadcastColumnMoved(t *testing.T) {
	pool := testPool(t)
	f := newFixture(t, pool)
	clearCaptured(t, pool)

	inActorTx(t, pool, f.userID, func(tx pgx.Tx) {
		if _, err := tx.Exec(context.Background(),
			`UPDATE public.columns SET position = 5, version = version + 1 WHERE id = $1`,
			f.sourceCol); err != nil {
			t.Fatalf("move column: %v", err)
		}
	})

	wantEnvelope(t, pool, f, "column_moved", f.sourceCol.String())
}

func TestBroadcastRollbackEmitsNothing(t *testing.T) {
	pool := testPool(t)
	f := newFixture(t, pool)
	clearCaptured(t, pool)

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if _, err := tx.Exec(ctx, `SELECT set_config('app.actor_id', $1, true)`, f.userID.String()); err != nil {
		t.Fatalf("set actor: %v", err)
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO public.tasks (id, column_id, title, position) VALUES ($1, $2, 'ghost', 0)`,
		uuid.New(), f.sourceCol); err != nil {
		t.Fatalf("insert task: %v", err)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("rollback: %v", err)
	}

	if events := capturedEvents(t, pool); len(events) != 0 {
		t.Fatalf("captured %d events after rollback, want none: %+v", len(events), events)
	}
}
