BEGIN;

-- ============================================================
-- 1. CREATE WORKSPACE
-- ============================================================

DO $$
DECLARE
    workspace_id UUID;
BEGIN

    INSERT INTO workspaces (
        id,
        name
    )
    VALUES (
        gen_random_uuid(),
        'NMC Project Management'
    )
    RETURNING id INTO workspace_id;


    -- ========================================================
    -- 2. CREATE WORKSPACE MEMBERSHIPS
    -- ========================================================

    INSERT INTO workspace_memberships (
        id,
        user_id,
        workspace_id,
        role
    )
    SELECT
        gen_random_uuid(),
        u.id,
        workspace_id,
        CASE
            WHEN u.email = 'admin@nmcpm.local' THEN 'owner'
            WHEN u.email = '11111@gmail.com' THEN 'admin'
            ELSE 'member'
        END
    FROM users u;


    -- ========================================================
    -- 3. CREATE BOARDS
    -- ========================================================

    INSERT INTO boards (
        id,
        workspace_id,
        name
    )
    VALUES
        (
            gen_random_uuid(),
            workspace_id,
            'NMC Product Development'
        ),
        (
            gen_random_uuid(),
            workspace_id,
            'NMC Marketing'
        );

END $$;

COMMIT;