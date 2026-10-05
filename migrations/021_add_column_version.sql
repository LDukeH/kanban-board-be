-- +goose Up

-- Optimistic concurrency control for columns.
--
-- Same contract as tasks.version (migration 020), applied to the board's column
-- ordering: PATCH /columns/{id}, PATCH /columns/{id}/position and
-- DELETE /columns/{id}?version=N must carry the version the client last read,
-- and the repository writes
--
--   UPDATE columns SET ..., version = version + 1 WHERE id = $x AND version = $y
--
-- so a guarded write that matched no row means either the column is gone
-- (404 Not Found) or another client already renamed or moved it
-- (409 Conflict).
--
-- Columns shifted by a reorder are bumped without their own version being
-- checked; only the destination column's version guards the write. The whole
-- shift runs in one transaction, so a stale destination rolls the entire
-- reorder back rather than committing part of it.
--
-- Deleting a column still cascades to its tasks through the existing
-- fk_tasks_column constraint. Those task rows disappear rather than being
-- reordered, so they carry no version to bump.
--
-- DEFAULT 1 backfills existing columns without a data migration, and creation
-- never sends a version.
--
-- IF NOT EXISTS keeps the migration safe to re-run, matching 019.
--
-- No index on version, for the same reason as tasks: it is only ever compared
-- on a row already narrowed by the primary key.
--
-- Rollback (this repo migrations are Up-only; run these by hand if needed):
--   ALTER TABLE columns DROP COLUMN IF EXISTS version;

ALTER TABLE columns
    ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1;