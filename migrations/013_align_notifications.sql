-- +goose Up

-- Align the notifications table with the Source of Truth (Notification contract, notifications section).
--
-- The previous shape (message + is_read boolean) is replaced with a generic schema
-- that supports future notification types without schema changes:
--
--   id, user_id, type, payload, read_at, created_at, updated_at
--
-- Existing rows are preserved. message and is_read are dropped, payload defaults
-- to an empty JSON object, read_at is nullable, updated_at is added.

ALTER TABLE notifications
    DROP COLUMN IF EXISTS message,
    DROP COLUMN IF EXISTS is_read,
    ADD COLUMN payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN read_at TIMESTAMPTZ,
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- Index for listing a user's notifications newest-first.
CREATE INDEX idx_notifications_user_id_created_at
    ON notifications (user_id, created_at DESC);