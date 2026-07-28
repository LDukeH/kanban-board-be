package workspace

import (
	"github.com/google/uuid"

	"time"
)

type Workspace struct {
	ID  uuid.UUID `json:"id"`
	Name string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// No boards yet because not implemented
}

type WorkspaceRole string

const(
	WorkSpaceRoleOwner WorkspaceRole = "owner"
	WorkspaceRoleAdmin WorkspaceRole = "admin"
	WorkspaceRoleMember WorkspaceRole = "member"
)

type WorkspaceMembership struct {
	ID uuid.UUID `json:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	UserID uuid.UUID `json:"user_id"`
	Role WorkspaceRole `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}