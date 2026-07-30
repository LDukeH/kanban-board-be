package board

import (
	"time"

	"github.com/google/uuid"
)

type Board struct {
    ID          uuid.UUID `json:"id"`
    WorkspaceID uuid.UUID `json:"workspace_id"`
    Name        string    `json:"name"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type BoardSummary struct {
    ID          uuid.UUID `json:"id"`
    WorkspaceID uuid.UUID `json:"workspace_id"`
    Name        string    `json:"name"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type BoardCreateRequest struct {
    Name string `json:"name"`
}

type BoardUpdateRequest struct {
    Name string `json:"name"`
}