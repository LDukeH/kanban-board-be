package column

import (
	"time"

	"github.com/google/uuid"
)

type Column struct {
	ID        uuid.UUID `json:"id"`
	BoardID   uuid.UUID `json:"board_id"`
	Name      string    `json:"name"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Version is the optimistic concurrency token (migration 021).
	Version int `json:"version"`
}

type ColumnCreateRequest struct {
	Name string `json:"name"`
}

// version is required, and is a pointer so an absent key is distinguishable from 0.
type ColumnUpdateRequest struct {
	Name    string `json:"name"`
	Version *int   `json:"version"`
}

// version is required, and is a pointer so an absent key is distinguishable from 0.
type ColumnPositionUpdateRequest struct {
	Position int  `json:"position"`
	Version  *int `json:"version"`
}
