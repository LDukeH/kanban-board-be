package task

import (
	"encoding/json"
	"fmt"
	"time"

	"kanban-board-be/internal/feature/label"

	"github.com/google/uuid"
)

type Task struct {
	ID       uuid.UUID `json:"id"`
	ColumnID uuid.UUID `json:"column_id"`
	Title    *string   `json:"title"`
	// Description is a Tiptap JSON document ({ "type": "doc", "content": [...] })
	Description *json.RawMessage `json:"description"`
	Position    int              `json:"position"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`

	Version int `json:"version"`

	Priority  *string            `json:"priority"`
	DueDate   Optional[DateOnly] `json:"due_date"`
	Completed *bool              `json:"completed"`

	Labels    []label.Label          `json:"labels,omitempty"`
	Assignees []TaskAssigneeResponse `json:"assignees,omitempty"`
}

const (
	PriorityNone   = "none"
	PriorityLow    = "low"
	PriorityMedium = "medium"
	PriorityHigh   = "high"
)

func isValidTaskPriority(p string) bool {
	switch p {
	case PriorityNone, PriorityLow, PriorityMedium, PriorityHigh:
		return true
	default:
		return false
	}
}

const DateLayout = "2006-01-02"

type DateOnly struct {
	time.Time
}

func (d DateOnly) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format(DateLayout))
}

func (d *DateOnly) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("due_date must be a string in %s format: %w", DateLayout, err)
	}

	parsed, err := time.Parse(DateLayout, raw)
	if err != nil {
		return fmt.Errorf("due_date must be in %s format: %w", DateLayout, err)
	}

	d.Time = parsed
	return nil
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

type TaskDescriptionRequest = json.RawMessage

type TaskCreateRequest struct {
	Title       *string                 `json:"title"`
	Description *TaskDescriptionRequest `json:"description"`
	Priority    *string                 `json:"priority"`
	// Absent or null both leave due_date NULL on create.
	DueDate   Optional[DateOnly] `json:"due_date"`
	Completed *bool              `json:"completed"`
}

type TaskUpdateRequest struct {
	Title       *string                 `json:"title"`
	Description *TaskDescriptionRequest `json:"description"`
	Priority    *string                 `json:"priority"`
	DueDate     Optional[DateOnly]      `json:"due_date"`
	Completed   *bool                   `json:"completed"`
	Version     *int                    `json:"version"`
}

func isValidTaskDescription(d json.RawMessage) bool {
	var header struct {
		Type string `json:"type"`
	}
	return json.Unmarshal(d, &header) == nil && header.Type == "doc"
}

type TaskPositionUpdateRequest struct {
	ColumnID uuid.UUID `json:"column_id"`
	Position int       `json:"position"`
	Version  *int      `json:"version"`
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
