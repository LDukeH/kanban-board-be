package realtime_test

import (
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func applyRealtimeMigration(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := testContext()

	for _, stmt := range []string{
		`CREATE EXTENSION IF NOT EXISTS pgcrypto`,
		`CREATE SCHEMA IF NOT EXISTS realtime`,
		`CREATE TABLE IF NOT EXISTS realtime.captured (
			topic text NOT NULL,
			event text NOT NULL,
			payload jsonb NOT NULL,
			created_at timestamptz NOT NULL DEFAULT now()
		)`,
		`CREATE OR REPLACE FUNCTION realtime.send(payload jsonb, event text, topic text, private boolean DEFAULT true)
		RETURNS void
		LANGUAGE plpgsql
		AS $stub$
		BEGIN
			INSERT INTO realtime.captured (topic, event, payload)
			VALUES (topic, event, payload);
			RETURN;
		END
		$stub$`,
		`CREATE TABLE IF NOT EXISTS public.users (
			id uuid PRIMARY KEY,
			email text NOT NULL UNIQUE,
			password_hash text
		)`,
		`CREATE TABLE IF NOT EXISTS public.workspaces (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			name text NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS public.boards (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			workspace_id uuid NOT NULL REFERENCES public.workspaces (id) ON DELETE CASCADE,
			name text NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS public.columns (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			board_id uuid NOT NULL REFERENCES public.boards (id) ON DELETE CASCADE,
			name text NOT NULL DEFAULT '',
			position integer NOT NULL,
			version integer NOT NULL DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS public.tasks (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			column_id uuid NOT NULL REFERENCES public.columns (id) ON DELETE CASCADE,
			title text NOT NULL DEFAULT '',
			description jsonb,
			position integer NOT NULL,
			version integer NOT NULL DEFAULT 1
		)`,
		`DROP TRIGGER IF EXISTS broadcast_tasks_changes ON public.tasks`,
		`DROP TRIGGER IF EXISTS broadcast_columns_changes ON public.columns`,
		`DROP FUNCTION IF EXISTS public.broadcast_board_event()`,
	} {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			t.Fatalf("test setup: %v", err)
		}
	}

	up := readMigrationUp(t, "../../migrations/023_realtime_board_broadcast.sql")
	for _, stmt := range splitStatements(up) {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			t.Fatalf("apply migration 023 statement: %v\n%s", err, stmt)
		}
	}
}

func readMigrationUp(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	src := string(raw)
	if idx := strings.Index(src, "-- +goose Down"); idx >= 0 {
		src = src[:idx]
	}
	return src
}

func splitStatements(src string) []string {
	var out []string
	var cur strings.Builder
	var dollarTag string
	inLineComment := false
	inBlockComment := false
	inSingle := false
	inDouble := false

	flush := func() {
		if meaningful(cur.String()) {
			out = append(out, cur.String())
		}
		cur.Reset()
	}

	for i := 0; i < len(src); {
		c := src[i]
		next := byte(0)
		if i+1 < len(src) {
			next = src[i+1]
		}

		if inLineComment {
			cur.WriteByte(c)
			if c == '\n' {
				inLineComment = false
			}
			i++
			continue
		}
		if inBlockComment {
			if c == '*' && next == '/' {
				cur.WriteString("*/")
				i += 2
				inBlockComment = false
				continue
			}
			cur.WriteByte(c)
			i++
			continue
		}
		if inSingle {
			if c == '\'' && next == '\'' {
				cur.WriteString("''")
				i += 2
				continue
			}
			cur.WriteByte(c)
			if c == '\'' {
				inSingle = false
			}
			i++
			continue
		}
		if inDouble {
			if c == '"' && next == '"' {
				cur.WriteString(`""`)
				i += 2
				continue
			}
			cur.WriteByte(c)
			if c == '"' {
				inDouble = false
			}
			i++
			continue
		}
		if dollarTag != "" {
			if strings.HasPrefix(src[i:], dollarTag) {
				cur.WriteString(dollarTag)
				i += len(dollarTag)
				dollarTag = ""
				continue
			}
			cur.WriteByte(c)
			i++
			continue
		}

		switch {
		case c == '-' && next == '-':
			inLineComment = true
			cur.WriteString("--")
			i += 2
		case c == '/' && next == '*':
			inBlockComment = true
			cur.WriteString("/*")
			i += 2
		case c == '\'':
			inSingle = true
			cur.WriteByte(c)
			i++
		case c == '"':
			inDouble = true
			cur.WriteByte(c)
			i++
		case c == '$':
			tag, ok := scanDollarTag(src, i)
			if !ok {
				cur.WriteByte(c)
				i++
				continue
			}
			dollarTag = tag
			cur.WriteString(tag)
			i += len(tag)
		case c == ';':
			cur.WriteByte(c)
			i++
			flush()
		default:
			cur.WriteByte(c)
			i++
		}
	}
	flush()
	return out
}

func scanDollarTag(src string, i int) (string, bool) {
	j := i + 1
	for j < len(src) && (src[j] == '_' ||
		(src[j] >= 'a' && src[j] <= 'z') ||
		(src[j] >= 'A' && src[j] <= 'Z') ||
		(src[j] >= '0' && src[j] <= '9')) {
		j++
	}
	if j < len(src) && src[j] == '$' {
		return src[i : j+1], true
	}
	return "", false
}

func meaningful(stmt string) bool {
	for _, line := range strings.Split(stmt, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		return true
	}
	return false
}
