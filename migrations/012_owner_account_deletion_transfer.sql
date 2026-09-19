-- +goose Up

-- Experimental: This migration is a bit more complex than the others, and it is not yet fully tested. It is recommended to run this migration in a staging environment first before applying it to production.
-- Not yet implemented


-- -- 012_owner_account_deletion_transfer.sql
-- -- Rule: when a user deletes their account:
-- --   * Workspaces they OWN are transferred to an existing admin (preferred),
-- --     else an existing member. If no one is left, the workspace is deleted.
-- --   * Workspaces where they were only a MEMBER are unaffected; only their
-- --     membership and task_assignees rows are voided.

-- CREATE OR REPLACE FUNCTION handle_user_deletion() RETURNS TRIGGER
-- LANGUAGE plpgsql
-- AS $$
-- DECLARE
--     owned_workspace_ids UUID[];
--     w_id                UUID;
--     successor_user_id   UUID;
-- BEGIN
--     -- 1) Snapshot the workspaces the deleted user owns.
--     SELECT array_agg(workspace_id)
--       INTO owned_workspace_ids
--       FROM workspace_memberships
--      WHERE user_id = OLD.id
--        AND role = 'owner';

--     -- 2) Demote the outgoing owner to 'member' so the one-owner-per-workspace
--     --    partial unique index is free for the successor.
--     UPDATE workspace_memberships
--        SET role = 'member', updated_at = NOW()
--      WHERE user_id = OLD.id
--        AND role = 'owner';

--     -- 3) Hand each owned workspace to an admin, else to a member.
--     FOREACH w_id IN ARRAY owned_workspace_ids LOOP
--         SELECT user_id
--           INTO successor_user_id
--           FROM workspace_memberships
--          WHERE workspace_id = w_id
--            AND user_id <> OLD.id
--            AND role IN ('admin', 'member')
--          ORDER BY (role = 'admin') DESC,
--                   created_at ASC,
--                   id ASC
--          LIMIT 1;

--         IF successor_user_id IS NOT NULL THEN
--             -- Transfer ownership to the successor.
--             UPDATE workspace_memberships
--                SET role = 'owner', updated_at = NOW()
--              WHERE workspace_id = w_id
--                AND user_id = successor_user_id;
--         ELSE
--             -- No admin/member left: the workspace cannot survive, delete it.
--             -- Boards -> columns -> tasks -> labels/assignees/memberships cascade.
--             DELETE FROM workspaces WHERE id = w_id;
--         END IF;
--     END LOOP;

--     -- 4) Void the leaving user's memberships and all task assignments.
--     DELETE FROM workspace_memberships WHERE user_id = OLD.id;
--     DELETE FROM task_assignees          WHERE user_id = OLD.id;

--     RETURN OLD;
-- END;
-- $$;

-- -- Idempotent re-run guard (no-op the FK cascades; rows already gone).
-- DROP TRIGGER IF EXISTS users_before_delete ON users;

-- CREATE TRIGGER users_before_delete
--     BEFORE DELETE ON users
--     FOR EACH ROW
--     EXECUTE FUNCTION handle_user_deletion();