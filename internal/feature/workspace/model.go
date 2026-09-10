package workspace

import (
	"time"

	"kanban-board-be/internal/feature/membership"

	"github.com/google/uuid"
)

type Workspace struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// No boards yet because not implemented
}

type WorkspaceSummary struct {
	ID        uuid.UUID                `json:"id"`
	Name      string                   `json:"name"`
	Role      membership.WorkspaceRole `json:"role"`
	CreatedAt time.Time                `json:"created_at"`
	UpdatedAt time.Time                `json:"updated_at"`
}

type WorkspaceCreateRequest struct {
	Name string `json:"name"`
}

type WorkspaceUpdateRequest struct {
	Name string `json:"name"`
}

type TransferOwnershipRequest struct {
	NewOwnerID uuid.UUID `json:"new_owner_id"`
}