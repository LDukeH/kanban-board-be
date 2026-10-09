package realtime_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("KANBAN_REALTIME_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("KANBAN_REALTIME_TEST_DATABASE_URL is not set")
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse KANBAN_REALTIME_TEST_DATABASE_URL: %v", err)
	}
	if !isDisposableDatabase(cfg.ConnConfig.Database) {
		t.Skipf("refusing to run against database %q: use a disposable test database", cfg.ConnConfig.Database)
	}

	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}
	t.Cleanup(pool.Close)

	var hasRealtime bool
	if err := pool.QueryRow(ctx, `SELECT to_regclass('realtime.messages') IS NOT NULL`).Scan(&hasRealtime); err != nil {
		t.Fatalf("probe realtime.messages: %v", err)
	}
	if hasRealtime {
		t.Skip("refusing to run: realtime.messages exists (this looks like a real Supabase project)")
	}

	applyRealtimeMigration(t, pool)
	return pool
}

func isDisposableDatabase(name string) bool {
	lowered := strings.ToLower(name)
	for _, marker := range []string{"test", "ci", "scratch", "ephemeral", "realtime"} {
		if strings.Contains(lowered, marker) {
			return true
		}
	}
	return false
}

type fixture struct {
	userID    uuid.UUID
	boardID   uuid.UUID
	sourceCol uuid.UUID
	targetCol uuid.UUID
}

func newFixture(t *testing.T, pool *pgxpool.Pool) fixture {
	t.Helper()
	ctx := context.Background()
	f := fixture{
		userID:    uuid.New(),
		boardID:   uuid.New(),
		sourceCol: uuid.New(),
		targetCol: uuid.New(),
	}
	workspaceID := uuid.New()

	if _, err := pool.Exec(ctx,
		`INSERT INTO public.users (id, email) VALUES ($1, $2)`,
		f.userID, f.userID.String()+"@example.com"); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if _, err := pool.Exec(ctx,
		`INSERT INTO public.workspaces (id, name) VALUES ($1, 'realtime test')`,
		workspaceID); err != nil {
		t.Fatalf("seed workspace: %v", err)
	}
	if _, err := pool.Exec(ctx,
		`INSERT INTO public.boards (id, workspace_id, name) VALUES ($1, $2, 'realtime test')`,
		f.boardID, workspaceID); err != nil {
		t.Fatalf("seed board: %v", err)
	}
	for i, col := range []uuid.UUID{f.sourceCol, f.targetCol} {
		if _, err := pool.Exec(ctx,
			`INSERT INTO public.columns (id, board_id, name, position) VALUES ($1, $2, $3, $4)`,
			col, f.boardID, "col", i); err != nil {
			t.Fatalf("seed column: %v", err)
		}
	}

	t.Cleanup(func() {
		ctx := context.Background()
		// Best effort: remove the fixture workspace (cascades to boards,
		// columns and tasks), the user and any captured stub events.
		_, _ = pool.Exec(ctx, `DELETE FROM public.workspaces WHERE id = $1`, workspaceID)
		_, _ = pool.Exec(ctx, `DELETE FROM public.users WHERE id = $1`, f.userID)
		_, _ = pool.Exec(ctx, `TRUNCATE realtime.captured`)
	})
	return f
}

// inActorTx runs fn inside one transaction with app.actor_id stamped, the way
// the repositories do after Begin, and commits.
func inActorTx(t *testing.T, pool *pgxpool.Pool, actor uuid.UUID, fn func(tx pgx.Tx)) {
	t.Helper()
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `SELECT set_config('app.actor_id', $1, true)`, actor.String()); err != nil {
		t.Fatalf("set actor: %v", err)
	}
	fn(tx)
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}
}

func clearCaptured(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), `TRUNCATE realtime.captured`); err != nil {
		t.Fatalf("clear captured: %v", err)
	}
}
