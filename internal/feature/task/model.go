package task

import (
	"time"

	"github.com/google/uuid"
)

type Task struct {
    ID          uuid.UUID `json:"id"`
    ColumnID    uuid.UUID `json:"column_id"`
    Title       *string    `json:"title"`
    Description *string   `json:"description"`
    Position    int       `json:"position"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type TaskAssignee struct {
    TaskID uuid.UUID `json:"task_id"`
    UserID uuid.UUID `json:"user_id"`
}

type TaskLabel struct {
    TaskID  uuid.UUID `json:"task_id"`
    LabelID uuid.UUID `json:"label_id"`
}

type TaskCreateRequest struct {
    Title       *string `json:"title"`
    Description *string `json:"description"`
}

type TaskUpdateRequest struct {
    Title       *string `json:"title"`
    Description *string `json:"description"`
}

type TaskPositionUpdateRequest struct {
    ColumnID uuid.UUID `json:"column_id"`
    Position int       `json:"position"`
}
