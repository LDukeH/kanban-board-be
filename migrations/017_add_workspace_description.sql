-- +goose Up

-- The Workspaces page renders a short description next to each workspace.
-- NOT NULL DEFAULT '' keeps the summary response a plain string: existing
-- workspaces read back as an empty description instead of NULL, so the Go
-- model needs no nullable handling and new workspaces need no extra write.

ALTER TABLE workspaces
    ADD COLUMN description TEXT NOT NULL DEFAULT '';
