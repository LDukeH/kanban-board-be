-- +goose Up

-- Optimistic concurrency control for tasks.
--
-- Every task carries a version counter that a client must echo back on every
-- mutation of an existing task: PATCH /tasks/{id}, PATCH /tasks/{id}/position
-- and DELETE /tasks/{id}?version=N. The repository writes
--
--   UPDATE tasks SET ..., version = version + 1 WHERE id = $x AND version = $y
--
-- so a guarded write that matched no row means either the task is gone
-- (404 Not Found) or another client already moved it on (409 Conflict).
--
-- Rows that a reorder shifts are bumped too, without their own version being
-- checked: a client holding a stale position for a neighbour refetches instead
-- of silently overwriting the reordering client.
--
-- DEFAULT 1 backfills existing rows without a data migration -- an existing
-- task starts at 1, which is exactly what a client sees on the next GET.
-- Creation never sends a version: the column default initialises it.
--
-- IF NOT EXISTS keeps the migration safe to re-run, matching the convention
-- established in 019_add_task_metadata.sql.
--
-- No index on version: it is only ever compared on a row already narrowed by
-- the primary key, so a separate index would never be used.
--
-- Rollback (this repo migrations are Up-only; run these by hand if needed):
--   ALTER TABLE tasks DROP COLUMN IF EXISTS version;

ALTER TABLE tasks
    ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1;