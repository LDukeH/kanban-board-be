-- +goose Up
-- +goose StatementBegin

-- Realtime board broadcast via Supabase Realtime Broadcast.
--
-- A row-level trigger on public.tasks and public.columns emits one event per
-- changed row into a private channel board:{board_id}. The Go layer never
-- publishes: it only sets app.actor_id inside the same transaction (see the
-- setActor helper), so the payload can name the acting user. Because
-- realtime.send writes to realtime.messages inside the caller's transaction,
-- an event is only dispatched on commit -- a rolled-back mutation leaves no
-- trace.
--
-- The function is SECURITY DEFINER with an empty search_path and every object
-- schema-qualified, matching the hardened style of the Supabase best-practice
-- guide. SECURITY DEFINER lets realtime.send insert into realtime.messages
-- even though the calling user would not have that privilege directly.
--
-- CI applies migrations on a bare postgres:16 image with no realtime schema,
-- so the RLS policy (which needs realtime.messages and auth.uid()) is guarded
-- by an existence check and is skipped there. The function and triggers are
-- created unconditionally: plpgsql late-binds realtime.send, so they compile
-- even when the schema is absent and simply never fire on a database that has
-- no Realtime extension enabled.

CREATE OR REPLACE FUNCTION public.broadcast_board_event()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
    v_board_id       uuid;
    v_event_type     text;
    v_entity_id      uuid;
    v_column_id      uuid;
    v_source_column  uuid;
    v_version        integer;
    v_payload        jsonb;
BEGIN
    IF TG_TABLE_NAME = 'tasks' THEN
        -- A task's board is reached through its column; NEW wins on
        -- INSERT/UPDATE, OLD on DELETE. When a column is deleted its tasks are
        -- cascaded away in the same transaction, so this lookup returns no row
        -- and the task event is skipped below (only column_deleted is emitted).
        v_entity_id := COALESCE(NEW.id, OLD.id);
        v_column_id := COALESCE(NEW.column_id, OLD.column_id);
        v_version   := COALESCE(NEW.version, OLD.version);

        SELECT c.board_id INTO v_board_id
        FROM public.columns AS c
        WHERE c.id = v_column_id;

        IF TG_OP = 'INSERT' THEN
            v_event_type    := 'task_created';
            v_source_column := NULL;
        ELSIF TG_OP = 'DELETE' THEN
            v_event_type    := 'task_deleted';
            v_source_column := NULL;
        ELSIF NEW.column_id IS DISTINCT FROM OLD.column_id
              OR NEW.position IS DISTINCT FROM OLD.position THEN
            v_event_type    := 'task_moved';
            v_source_column := OLD.column_id;
        ELSE
            v_event_type    := 'task_updated';
            v_source_column := NULL;
        END IF;
    ELSIF TG_TABLE_NAME = 'columns' THEN
        v_entity_id     := COALESCE(NEW.id, OLD.id);
        v_column_id     := v_entity_id;
        v_version       := COALESCE(NEW.version, OLD.version);
        v_board_id      := COALESCE(NEW.board_id, OLD.board_id);
        v_source_column := NULL;

        IF TG_OP = 'INSERT' THEN
            v_event_type := 'column_created';
        ELSIF TG_OP = 'DELETE' THEN
            v_event_type := 'column_deleted';
        ELSIF NEW.position IS DISTINCT FROM OLD.position THEN
            v_event_type := 'column_moved';
        ELSE
            v_event_type := 'column_updated';
        END IF;
    ELSE
        RETURN NULL;
    END IF;

    -- No resolvable board (cascaded task, orphaned row): stay silent.
    IF v_board_id IS NULL THEN
        RETURN NULL;
    END IF;

    v_payload := jsonb_build_object(
        'id',               gen_random_uuid(),
        'type',             v_event_type,
        'board_id',         v_board_id,
        'entity_id',        v_entity_id,
        'column_id',        v_column_id,
        'source_column_id', v_source_column,
        'version',          v_version,
        'actor_id',         nullif(current_setting('app.actor_id', true), ''),
        'occurred_at',      now()
    );

    PERFORM realtime.send(
        v_payload,
        v_event_type,
        'board:' || v_board_id::text,
        true
    );

    RETURN NULL;
END;
$$;

-- +goose StatementEnd

DROP TRIGGER IF EXISTS broadcast_tasks_changes ON public.tasks;

CREATE TRIGGER broadcast_tasks_changes
AFTER INSERT OR UPDATE OR DELETE ON public.tasks
FOR EACH ROW
EXECUTE FUNCTION public.broadcast_board_event();

DROP TRIGGER IF EXISTS broadcast_columns_changes ON public.columns;

CREATE TRIGGER broadcast_columns_changes
AFTER INSERT OR UPDATE OR DELETE ON public.columns
FOR EACH ROW
EXECUTE FUNCTION public.broadcast_board_event();

-- +goose StatementBegin

-- Read authorization for the private board channel. Skipped when the realtime
-- or auth schema is absent (bare CI database). A member of the board's
-- workspace may read broadcast messages whose topic is that board; nobody
-- else can. No INSERT policy is created: only the server publishes.
DO $$
BEGIN
    IF to_regclass('realtime.messages') IS NOT NULL
       AND to_regprocedure('auth.uid()') IS NOT NULL
       AND to_regprocedure('realtime.topic()') IS NOT NULL THEN
        DROP POLICY IF EXISTS "board broadcast read" ON realtime.messages;
        CREATE POLICY "board broadcast read"
        ON realtime.messages
        FOR SELECT
        TO authenticated
        USING (
            extension = 'broadcast'
            AND EXISTS (
                SELECT 1
                FROM public.boards AS b
                JOIN public.workspace_memberships AS wm
                    ON wm.workspace_id = b.workspace_id
                WHERE wm.user_id = (SELECT auth.uid())
                  AND 'board:' || b.id::text = (SELECT realtime.topic())
            )
        );
    END IF;
END $$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DO $$
BEGIN
    IF to_regclass('realtime.messages') IS NOT NULL THEN
        DROP POLICY IF EXISTS "board broadcast read" ON realtime.messages;
    END IF;
END $$;

-- +goose StatementEnd

DROP TRIGGER IF EXISTS broadcast_tasks_changes ON public.tasks;
DROP TRIGGER IF EXISTS broadcast_columns_changes ON public.columns;
DROP FUNCTION IF EXISTS public.broadcast_board_event();
