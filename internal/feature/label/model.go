package label

import (
	"time"

	"github.com/google/uuid"
)

type Label struct {
	ID        uuid.UUID `json:"id"`
	BoardID   uuid.UUID `json:"board_id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TaskLabel struct {
	TaskID    uuid.UUID `json:"task_id"`
	LabelID   uuid.UUID `json:"label_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LabelCreateRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type LabelUpdateRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}
