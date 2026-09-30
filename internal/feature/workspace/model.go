package workspace

import (
	"time"

	"kanban-board-be/internal/feature/membership"

	"github.com/google/uuid"
)

type Workspace struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// No boards yet because not implemented
}

// List projection for GET /workspaces; counts use subqueries to avoid a cartesian product.
type WorkspaceSummary struct {
	ID          uuid.UUID                `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Role        membership.WorkspaceRole `json:"role"`
	MemberCount int                      `json:"member_count"`
	BoardCount  int                      `json:"board_count"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
}

// Column is unbounded TEXT, so the limit is enforced in Go only.
const maxWorkspaceDescriptionLength = 2000

// Body of POST /workspaces. Description is optional; omitted stores "".
type WorkspaceCreateRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

// Body of PATCH /workspaces/{workspaceID}. Nil description = unchanged, "" = cleared.
type WorkspaceUpdateRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

// Partial-update payload: a nil field means "leave unchanged".
type WorkspaceUpdate struct {
	ID          uuid.UUID
	Name        *string
	Description *string
}

type TransferOwnershipRequest struct {
	NewOwnerID uuid.UUID `json:"new_owner_id"`
}
