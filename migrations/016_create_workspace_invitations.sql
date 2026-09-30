-- +goose Up

-- Workspace invitations are a separate domain entity from memberships.
-- A pending invitation grants no workspace access; access is granted only
-- when the invitee accepts and a workspace_memberships row is created in
-- the same transaction as the acceptance.
--
-- Lifecycle: pending -> accepted | declined | cancelled | expired.
-- Only pending rows may transition. Expiry is enforced by the service via
-- expires_at, which also converges stale pending rows to 'expired'.

CREATE TABLE workspace_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    workspace_id UUID NOT NULL,
    invitee_user_id UUID NOT NULL,
    invited_by_user_id UUID NOT NULL,

    status TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '7 days'),

    CONSTRAINT fk_workspace_invitations_workspace
        FOREIGN KEY (workspace_id)
        REFERENCES workspaces(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_workspace_invitations_invitee_user
        FOREIGN KEY (invitee_user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_workspace_invitations_invited_by_user
        FOREIGN KEY (invited_by_user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT chk_workspace_invitations_status
        CHECK (status IN ('pending', 'accepted', 'declined', 'cancelled', 'expired'))
);

-- At most one pending invitation per (workspace, invitee). Terminal states
-- free the slot, so a user can be re-invited after a previous invitation was
-- declined/cancelled/expired/accepted.
CREATE UNIQUE INDEX uq_workspace_invitations_pending
ON workspace_invitations (workspace_id, invitee_user_id)
WHERE status = 'pending';

CREATE INDEX idx_workspace_invitations_workspace_id
ON workspace_invitations (workspace_id);

CREATE INDEX idx_workspace_invitations_invitee_user_id
ON workspace_invitations (invitee_user_id);