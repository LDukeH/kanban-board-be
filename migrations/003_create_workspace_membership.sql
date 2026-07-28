CREATE TABLE workspace_memberships (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    workspace_id UUID NOT NULL,
    role TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_workspace_memberships_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_workspace_memberships_workspace
        FOREIGN KEY (workspace_id)
        REFERENCES workspaces(id)
        ON DELETE CASCADE,

    CONSTRAINT uq_workspace_memberships_user_workspace
        UNIQUE (user_id, workspace_id), 

    CONSTRAINT chk_workspace_memberships_role
        CHECK (role IN ('owner', 'admin', 'member'))
);

CREATE UNIQUE INDEX uq_workspace_memberships_one_owner
ON workspace_memberships (workspace_id)
WHERE role = 'owner';