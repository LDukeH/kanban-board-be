-- +goose Up

-- Task planning metadata: priority, due date and completion state.
--
-- These three fields back the "task planning metadata + filtering" feature. The
-- frontend filters entirely client-side (src/lib/taskFilter.ts) over the tasks it
-- has already fetched from GET /columns/{id}/tasks, so no new query endpoint is
-- required -- the fields only have to appear in the task response.
--
-- Defaults keep every existing row valid without a backfill:
--   priority  = none   (neutral value; the CHECK below keeps the set closed)
--   due_date  = NULL   (no due date)
--   completed = false  (nothing was complete before this feature existed)
--
-- IF NOT EXISTS / the guarded constraint keep the migration safe to re-run,
-- matching the convention established in 018_perf_indexes.sql.
--
-- Rollback (this repo migrations are Up-only; run these by hand if needed):
--   DROP INDEX IF EXISTS idx_tasks_column_id_due_date;
--   ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_priority_check;
--   ALTER TABLE tasks DROP COLUMN IF EXISTS completed;
--   ALTER TABLE tasks DROP COLUMN IF EXISTS due_date;
--   ALTER TABLE tasks DROP COLUMN IF EXISTS priority;

ALTER TABLE tasks
    ADD COLUMN IF NOT EXISTS priority TEXT NOT NULL DEFAULT 'none',
    ADD COLUMN IF NOT EXISTS due_date DATE,
    ADD COLUMN IF NOT EXISTS completed BOOLEAN NOT NULL DEFAULT false;

-- Priority is a closed set. Postgres has no ADD CONSTRAINT IF NOT EXISTS, so
-- guard it explicitly. Goose splits statements on ";", so the dollar-quoted block
-- is wrapped in StatementBegin/StatementEnd to be sent as one statement.
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'tasks_priority_check'
    ) THEN
        ALTER TABLE tasks
            ADD CONSTRAINT tasks_priority_check
                CHECK (priority IN ('none', 'low', 'medium', 'high'));
    END IF;
END $$;
-- +goose StatementEnd

-- Serves the board due-date filter/sort within a column, mirroring the
-- (column_id, ...) leading column of idx_tasks_column_id_position.
CREATE INDEX IF NOT EXISTS idx_tasks_column_id_due_date
    ON tasks (column_id, due_date);