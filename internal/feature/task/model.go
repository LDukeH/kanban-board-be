package task

import (
	"encoding/json"
	"time"

	"kanban-board-be/internal/feature/label"

	"github.com/google/uuid"
)

type Task struct {
	ID       uuid.UUID `json:"id"`
	ColumnID uuid.UUID `json:"column_id"`
	Title    *string   `json:"title"`
	// Description is a Tiptap JSON document ({ "type": "doc", "content": [...] })
	// stored in a JSONB column. nil (JSON null) means no description.
	Description *json.RawMessage `json:"description"`
	Position    int              `json:"position"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`

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

// TaskDescriptionRequest is a Tiptap JSON document sent by the frontend
// ({ "type": "doc", "content": [...] }). nil = field absent (no change on
// update / no description on create). Validated by isValidTaskDescription.
type TaskDescriptionRequest = json.RawMessage

type TaskCreateRequest struct {
	Title       *string                 `json:"title"`
	Description *TaskDescriptionRequest `json:"description"`
}

type TaskUpdateRequest struct {
	Title       *string                 `json:"title"`
	Description *TaskDescriptionRequest `json:"description"`
}

// isValidTaskDescription checks that the raw JSON is a Tiptap document with a
// top-level "type": "doc". The JSONB column also enforces valid JSON, but this
// rejects non-doc JSON early with a clear 400.
func isValidTaskDescription(d json.RawMessage) bool {
	var header struct {
		Type string `json:"type"`
	}
	return json.Unmarshal(d, &header) == nil && header.Type == "doc"
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
