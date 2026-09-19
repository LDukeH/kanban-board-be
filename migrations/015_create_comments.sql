-- +goose Up

-- Comments belong to a task and are authored by an authenticated user.
--
-- A comment is immutable except deletion: there is no update endpoint, so
-- updated_at never changes in practice. It is kept only for model consistency
-- with the other tables in this codebase (do not build an edit feature around it).
--
-- The workspace is not stored on the comment; it is resolved through
--   comment.task_id -> task.board_id -> board.workspace_id
-- and workspace membership is the authorization boundary.

CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    task_id UUID NOT NULL
        REFERENCES tasks(id)
        ON DELETE CASCADE,

    author_id UUID NOT NULL
        REFERENCES users(id)
        ON DELETE CASCADE,

    content TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Support listing a task's comments oldest -> newest without a sort.
CREATE INDEX idx_comments_task_id_created_at
    ON comments (task_id, created_at);