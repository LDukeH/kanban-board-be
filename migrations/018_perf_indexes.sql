-- +goose Up

-- Performance indexes for the read paths the API actually issues.
--
-- Every foreign key in this schema was declared without an index. PostgreSQL
-- does not create one automatically for the referencing side of a FK, so the
-- board page (boards -> columns -> tasks -> labels/assignees) performed
-- sequential scans per column and per board. The composite indexes below are
-- ordered so one index both filters and satisfies the ORDER BY of its endpoint,
-- which also removes the sort step.
--
-- Indexes intentionally omitted because an existing constraint already covers
-- the leading column:
--   labels (board_id, name)                  -> uq_labels_board_name
--   task_labels (task_id, ...)               -> pk_task_labels
--   task_assignees (task_id, ...)            -> pk_task_assignees
--   workspace_memberships (user_id, ...)     -> uq_workspace_memberships_user_workspace
--   comments (task_id, created_at)           -> idx_comments_task_id_created_at
--   notifications (user_id, created_at DESC) -> idx_notifications_user_id_created_at
--
-- IF NOT EXISTS keeps the migration safe to re-run.
--
-- NOTE: plain CREATE INDEX takes a SHARE lock that blocks writes for the
-- duration of the build. That is fine for this dataset; for a large table run
-- the statements manually with CREATE INDEX CONCURRENTLY and then record this
-- migration as applied.

-- GET /workspaces/{id}/boards, plus the board_count aggregate on GET /workspaces.
CREATE INDEX IF NOT EXISTS idx_boards_workspace_id
    ON boards (workspace_id);

-- GET /boards/{id}/columns (filter + ORDER BY position).
-- Also serves the FK cascade from boards.
CREATE INDEX IF NOT EXISTS idx_columns_board_id_position
    ON columns (board_id, position);

-- GET /columns/{id}/tasks (filter + ORDER BY position) and the MAX(position)
-- lookup in task creation. Also serves the FK cascade from columns.
CREATE INDEX IF NOT EXISTS idx_tasks_column_id_position
    ON tasks (column_id, position);

-- Membership list per workspace, member_count on GET /workspaces, and the CTE
-- join in task.GetByColumnID. The existing partial unique index on
-- workspace_memberships (workspace_id) only covers rows where role = 'owner',
-- so it cannot serve these reads.
CREATE INDEX IF NOT EXISTS idx_workspace_memberships_workspace_id
    ON workspace_memberships (workspace_id);

-- DELETE /workspaces/{id}/members/{membershipID} removes the user's task
-- assignments across a workspace; account deletion cascades the same way. The
-- primary key is (task_id, user_id), so user_id lookups had no index.
CREATE INDEX IF NOT EXISTS idx_task_assignees_user_id
    ON task_assignees (user_id);

-- Deleting a label cascades into task_labels; the primary key is
-- (task_id, label_id), so label_id lookups had no index.
CREATE INDEX IF NOT EXISTS idx_task_labels_label_id
    ON task_labels (label_id);

-- Comment deletion and account deletion cascade on author_id.
CREATE INDEX IF NOT EXISTS idx_comments_author_id
    ON comments (author_id);

-- Invited-by account deletion cascades on invited_by_user_id.
CREATE INDEX IF NOT EXISTS idx_workspace_invitations_invited_by_user_id
    ON workspace_invitations (invited_by_user_id);

-- GET /notifications/unread-count and GET /notifications?unread_only=true.
-- Both read only unread rows, so index just those: the partial index stays
-- small regardless of how many notifications are already read.
CREATE INDEX IF NOT EXISTS idx_notifications_user_id_unread
    ON notifications (user_id, created_at DESC)
    WHERE read_at IS NULL;
