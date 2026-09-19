BEGIN;

DO $$
DECLARE
    -- Workspace
    v_workspace_id UUID;

    -- Boards
    v_dev_board_id UUID;
    v_marketing_board_id UUID;

    -- Development columns
    v_dev_backlog_id UUID;
    v_dev_todo_id UUID;
    v_dev_progress_id UUID;
    v_dev_review_id UUID;
    v_dev_done_id UUID;

    -- Marketing columns
    v_marketing_ideas_id UUID;
    v_marketing_planned_id UUID;
    v_marketing_progress_id UUID;
    v_marketing_done_id UUID;

    -- Tasks
    v_task_api_id UUID;
    v_task_auth_id UUID;
    v_task_ui_id UUID;
    v_task_testing_id UUID;

    v_task_campaign_id UUID;
    v_task_content_id UUID;
    v_task_social_id UUID;

    -- Labels
    v_bug_label_id UUID;
    v_feature_label_id UUID;
    v_backend_label_id UUID;
    v_frontend_label_id UUID;
    v_urgent_label_id UUID;

    v_marketing_label_id UUID;
    v_content_label_id UUID;
    v_campaign_label_id UUID;

    -- Users
    v_owner_id UUID;
    v_admin_id UUID;
    v_alice_id UUID;
    v_bob_id UUID;
    v_user_123456_id UUID;
    v_user_12456_id UUID;

BEGIN

    -- ========================================================
    -- 1. GET WORKSPACE
    -- ========================================================

    SELECT id
    INTO v_workspace_id
    FROM workspaces
    WHERE name = 'NMC Project Management'
    LIMIT 1;


    IF v_workspace_id IS NULL THEN
        RAISE EXCEPTION 'Workspace "NMC Project Management" not found';
    END IF;


    -- ========================================================
    -- 2. GET BOARDS
    -- ========================================================

    SELECT id
    INTO v_dev_board_id
    FROM boards
    WHERE workspace_id = v_workspace_id
      AND name = 'NMC Product Development'
    LIMIT 1;

    SELECT id
    INTO v_marketing_board_id
    FROM boards
    WHERE workspace_id = v_workspace_id
      AND name = 'NMC Marketing'
    LIMIT 1;


    IF v_dev_board_id IS NULL OR v_marketing_board_id IS NULL THEN
        RAISE EXCEPTION 'Required boards not found';
    END IF;


    -- ========================================================
    -- 3. GET USERS
    -- ========================================================

    SELECT id INTO v_owner_id
    FROM users
    WHERE email = 'admin@nmcpm.local';

    SELECT id INTO v_admin_id
    FROM users
    WHERE email = '11111@gmail.com';

    SELECT id INTO v_alice_id
    FROM users
    WHERE email = 'alice@example.com';

    SELECT id INTO v_bob_id
    FROM users
    WHERE email = 'bob@example.com';

    SELECT id INTO v_user_123456_id
    FROM users
    WHERE email = '123456';

    SELECT id INTO v_user_12456_id
    FROM users
    WHERE email = '12456';


    -- ========================================================
    -- 4. CREATE DEVELOPMENT BOARD COLUMNS
    -- ========================================================

    INSERT INTO columns (
        id,
        board_id,
        name,
        position
    )
    VALUES
        (gen_random_uuid(), v_dev_board_id, 'Backlog', 0)
    RETURNING id INTO v_dev_backlog_id;

    INSERT INTO columns (
        id,
        board_id,
        name,
        position
    )
    VALUES
        (gen_random_uuid(), v_dev_board_id, 'To Do', 1)
    RETURNING id INTO v_dev_todo_id;

    INSERT INTO columns (
        id,
        board_id,
        name,
        position
    )
    VALUES
        (gen_random_uuid(), v_dev_board_id, 'In Progress', 2)
    RETURNING id INTO v_dev_progress_id;

    INSERT INTO columns (
        id,
        board_id,
        name,
        position
    )
    VALUES
        (gen_random_uuid(), v_dev_board_id, 'Review', 3)
    RETURNING id INTO v_dev_review_id;

    INSERT INTO columns (
        id,
        board_id,
        name,
        position
    )
    VALUES
        (gen_random_uuid(), v_dev_board_id, 'Done', 4)
    RETURNING id INTO v_dev_done_id;


    -- ========================================================
    -- 5. CREATE MARKETING BOARD COLUMNS
    -- ========================================================

    INSERT INTO columns (
        id,
        board_id,
        name,
        position
    )
    VALUES
        (gen_random_uuid(), v_marketing_board_id, 'Ideas', 0)
    RETURNING id INTO v_marketing_ideas_id;

    INSERT INTO columns (
        id,
        board_id,
        name,
        position
    )
    VALUES
        (gen_random_uuid(), v_marketing_board_id, 'Planned', 1)
    RETURNING id INTO v_marketing_planned_id;

    INSERT INTO columns (
        id,
        board_id,
        name,
        position
    )
    VALUES
        (gen_random_uuid(), v_marketing_board_id, 'In Progress', 2)
    RETURNING id INTO v_marketing_progress_id;

    INSERT INTO columns (
        id,
        board_id,
        name,
        position
    )
    VALUES
        (gen_random_uuid(), v_marketing_board_id, 'Done', 3)
    RETURNING id INTO v_marketing_done_id;


    -- ========================================================
    -- 6. CREATE DEVELOPMENT TASKS
    -- ========================================================

    INSERT INTO tasks (
        id,
        column_id,
        title,
        description,
        position
    )
    VALUES
        (
            gen_random_uuid(),
            v_dev_backlog_id,
            'Implement Task Assignment API',
            'Create API endpoints for assigning and unassigning users from tasks.',
            0
        )
    RETURNING id INTO v_task_api_id;

    INSERT INTO tasks (
        id,
        column_id,
        title,
        description,
        position
    )
    VALUES
        (
            gen_random_uuid(),
            v_dev_todo_id,
            'Implement JWT Authentication',
            'Implement login authentication and JWT middleware for protected routes.',
            0
        )
    RETURNING id INTO v_task_auth_id;

    INSERT INTO tasks (
        id,
        column_id,
        title,
        description,
        position
    )
    VALUES
        (
            gen_random_uuid(),
            v_dev_progress_id,
            'Build Kanban Board UI',
            'Create the frontend board interface with columns and draggable tasks.',
            0
        )
    RETURNING id INTO v_task_ui_id;

    INSERT INTO tasks (
        id,
        column_id,
        title,
        description,
        position
    )
    VALUES
        (
            gen_random_uuid(),
            v_dev_review_id,
            'Write Integration Tests',
            'Test workspace, board, task and assignment workflows.',
            0
        )
    RETURNING id INTO v_task_testing_id;


    -- ========================================================
    -- 7. CREATE MARKETING TASKS
    -- ========================================================

    INSERT INTO tasks (
        id,
        column_id,
        title,
        description,
        position
    )
    VALUES
        (
            gen_random_uuid(),
            v_marketing_ideas_id,
            'Plan Product Launch Campaign',
            'Define the strategy and goals for the upcoming product launch.',
            0
        )
    RETURNING id INTO v_task_campaign_id;

    INSERT INTO tasks (
        id,
        column_id,
        title,
        description,
        position
    )
    VALUES
        (
            gen_random_uuid(),
            v_marketing_planned_id,
            'Create Product Announcement',
            'Write the announcement content for the new product release.',
            0
        )
    RETURNING id INTO v_task_content_id;

    INSERT INTO tasks (
        id,
        column_id,
        title,
        description,
        position
    )
    VALUES
        (
            gen_random_uuid(),
            v_marketing_progress_id,
            'Prepare Social Media Posts',
            'Prepare launch posts for the company social media channels.',
            0
        )
    RETURNING id INTO v_task_social_id;


    -- ========================================================
    -- 8. CREATE DEVELOPMENT LABELS
    -- ========================================================

    INSERT INTO labels (
        id,
        board_id,
        name,
        color
    )
    VALUES
        (gen_random_uuid(), v_dev_board_id, 'Bug', '#e5484d')
    RETURNING id INTO v_bug_label_id;

    INSERT INTO labels (
        id,
        board_id,
        name,
        color
    )
    VALUES
        (gen_random_uuid(), v_dev_board_id, 'Feature', '#30a46c')
    RETURNING id INTO v_feature_label_id;

    INSERT INTO labels (
        id,
        board_id,
        name,
        color
    )
    VALUES
        (gen_random_uuid(), v_dev_board_id, 'Backend', '#3e63dd')
    RETURNING id INTO v_backend_label_id;

    INSERT INTO labels (
        id,
        board_id,
        name,
        color
    )
    VALUES
        (gen_random_uuid(), v_dev_board_id, 'Frontend', '#8e4ec6')
    RETURNING id INTO v_frontend_label_id;

    INSERT INTO labels (
        id,
        board_id,
        name,
        color
    )
    VALUES
        (gen_random_uuid(), v_dev_board_id, 'Urgent', '#f76b15')
    RETURNING id INTO v_urgent_label_id;


    -- ========================================================
    -- 9. CREATE MARKETING LABELS
    -- ========================================================

    INSERT INTO labels (
        id,
        board_id,
        name,
        color
    )
    VALUES
        (gen_random_uuid(), v_marketing_board_id, 'Marketing', '#e5484d')
    RETURNING id INTO v_marketing_label_id;

    INSERT INTO labels (
        id,
        board_id,
        name,
        color
    )
    VALUES
        (gen_random_uuid(), v_marketing_board_id, 'Content', '#30a46c')
    RETURNING id INTO v_content_label_id;

    INSERT INTO labels (
        id,
        board_id,
        name,
        color
    )
    VALUES
        (gen_random_uuid(), v_marketing_board_id, 'Campaign', '#3e63dd')
    RETURNING id INTO v_campaign_label_id;


    -- ========================================================
    -- 10. ASSIGN USERS TO DEVELOPMENT TASKS
    -- ========================================================

    INSERT INTO task_assignees (task_id, user_id)
    VALUES
        (v_task_api_id, v_admin_id),
        (v_task_api_id, v_owner_id),

        (v_task_auth_id, v_owner_id),

        (v_task_ui_id, v_alice_id),
        (v_task_ui_id, v_bob_id),

        (v_task_testing_id, v_admin_id);


    -- ========================================================
    -- 11. ASSIGN USERS TO MARKETING TASKS
    -- ========================================================

    INSERT INTO task_assignees (task_id, user_id)
    VALUES
        (v_task_campaign_id, v_owner_id),

        (v_task_content_id, v_alice_id),

        (v_task_social_id, v_bob_id),
        (v_task_social_id, v_user_123456_id);


    -- ========================================================
    -- 12. ATTACH LABELS TO DEVELOPMENT TASKS
    -- ========================================================

    INSERT INTO task_labels (task_id, label_id)
    VALUES
        (v_task_api_id, v_feature_label_id),
        (v_task_api_id, v_backend_label_id),

        (v_task_auth_id, v_feature_label_id),
        (v_task_auth_id, v_backend_label_id),
        (v_task_auth_id, v_urgent_label_id),

        (v_task_ui_id, v_feature_label_id),
        (v_task_ui_id, v_frontend_label_id),

        (v_task_testing_id, v_backend_label_id);


    -- ========================================================
    -- 13. ATTACH LABELS TO MARKETING TASKS
    -- ========================================================

    INSERT INTO task_labels (task_id, label_id)
    VALUES
        (v_task_campaign_id, v_campaign_label_id),
        (v_task_campaign_id, v_marketing_label_id),

        (v_task_content_id, v_content_label_id),
        (v_task_content_id, v_marketing_label_id),

        (v_task_social_id, v_content_label_id),
        (v_task_social_id, v_marketing_label_id);

END $$;

COMMIT;