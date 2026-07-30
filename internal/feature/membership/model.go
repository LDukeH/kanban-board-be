package membership

import (
	"time"

	"github.com/google/uuid"
)

type WorkspaceRole string

const(
	Owner WorkspaceRole = "owner"
	Admin WorkspaceRole = "admin"
	Member WorkspaceRole = "member"
)

type WorkspaceMembership struct {
	ID uuid.UUID `json:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	UserID uuid.UUID `json:"user_id"`
	Role WorkspaceRole `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MembershipCreateRequest struct {
	UserID uuid.UUID `json:"user_id"`
	Role WorkspaceRole `json:"role"`
}

type ChangeRoleRequest struct {
	Role WorkspaceRole `json:"role"`
}
