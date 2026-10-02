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
	// stored in a JSONB column. nil (JSON null) means no description.
	Description *json.RawMessage `json:"description"`
	Position    int              `json:"position"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`

	// Priority is one of PriorityNone/PriorityLow/PriorityMedium/PriorityHigh.
	// The column is NOT NULL DEFAULT 'none', so a response always carries a value.
	Priority *string `json:"priority"`
	// DueDate is tri-state on the request path (absent = unchanged, null = clear,
	// value = set) and always Value or Null on the response path. The column is a
	// nullable DATE, so the value is a calendar date with no time component.
	DueDate Optional[DateOnly] `json:"due_date"`
	// Completed is the task's done flag. The column is NOT NULL DEFAULT false.
	Completed *bool `json:"completed"`

	// This doesn't exist in the database, but is used to return the labels and assignees associated with the task
	Labels    []label.Label          `json:"labels,omitempty"`
	Assignees []TaskAssigneeResponse `json:"assignees,omitempty"`
}

// Priority values accepted by the tasks_priority_check constraint added in
// migration 019. Kept as constants so handlers, the service and tests share one
// definition of the closed set.
const (
	PriorityNone   = "none"
	PriorityLow    = "low"
	PriorityMedium = "medium"
	PriorityHigh   = "high"
)

// isValidTaskPriority reports whether p is one of the values the database's
// tasks_priority_check constraint allows. The service turns a false into
// apperrors.ErrValidation so a caller gets a 400 instead of a 500.
func isValidTaskPriority(p string) bool {
	switch p {
	case PriorityNone, PriorityLow, PriorityMedium, PriorityHigh:
		return true
	default:
		return false
	}
}

// DateLayout is the only accepted due_date wire format. It matches the
// PostgreSQL DATE column and the HTML date input the frontend renders, rather
// than the RFC 3339 timestamp time.Time would otherwise emit.
const DateLayout = "2006-01-02"

// DateOnly is a calendar date without a time component, serialized as
// "YYYY-MM-DD".
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

// TaskDescriptionRequest is a Tiptap JSON document sent by the frontend
// ({ "type": "doc", "content": [...] }). nil = field absent (no change on
// update / no description on create). Validated by isValidTaskDescription.
type TaskDescriptionRequest = json.RawMessage

type TaskCreateRequest struct {
	Title       *string                 `json:"title"`
	Description *TaskDescriptionRequest `json:"description"`
	Priority    *string                 `json:"priority"`
	// Absent or null both leave due_date NULL on create.
	DueDate   Optional[DateOnly] `json:"due_date"`
	Completed *bool              `json:"completed"`
}

// TaskUpdateRequest is a partial update: every field absent means "leave the
// stored value alone". due_date is the exception that needs three states, since
// JSON null clears it while an absent key does not (see Optional).
type TaskUpdateRequest struct {
	Title       *string                 `json:"title"`
	Description *TaskDescriptionRequest `json:"description"`
	Priority    *string                 `json:"priority"`
	DueDate     Optional[DateOnly]      `json:"due_date"`
	Completed   *bool                   `json:"completed"`
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
