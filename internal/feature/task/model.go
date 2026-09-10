package task

import (
	"time"

	"kanban-board-be/internal/feature/label"

	"github.com/google/uuid"
)

type Task struct {
	ID          uuid.UUID `json:"id"`
	ColumnID    uuid.UUID `json:"column_id"`
	Title       *string   `json:"title"`
	Description *string   `json:"description"`
	Position    int       `json:"position"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// This doesn't exist in the database, but is used to return the labels and assignees associated with the task
	Labels    []label.Label          `json:"labels,omitempty"`
	Assignees []TaskAssigneeResponse `json:"assignees,omitempty"`
}

type TaskAssignee struct {
	TaskID uuid.UUID `json:"task_id"`
	UserID uuid.UUID `json:"user_id"`
}

type TaskAssigneeResponse struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
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

type TaskLabelCreateRequest struct {
	LabelID uuid.UUID `json:"label_id"`
}

type TaskAssignUserRequest struct {
	UserID uuid.UUID `json:"user_id"`
}

type TaskRemoveUserRequest struct {
	UserID uuid.UUID `json:"user_id"`
}

type UsersTaskResponse struct {
	Email  string    `json:"email"`
	UserID uuid.UUID `json:"user_id"`
}
