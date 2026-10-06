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

type ColumnUpdateRequest struct {
	Name    string `json:"name"`
	Version *int   `json:"version"`
}

type ColumnPositionUpdateRequest struct {
	Position int  `json:"position"`
	Version  *int `json:"version"`
}
