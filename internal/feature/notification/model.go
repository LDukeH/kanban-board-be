package notification

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	TypeTaskAssigned     NotificationType = "task_assigned"
	TypeWorkspaceInvited NotificationType = "workspace_invited"
)

type Notification struct {
	ID     uuid.UUID        `json:"id"`
	UserID uuid.UUID        `json:"user_id"`
	Type   NotificationType `json:"type"`
	// Payload is display/navigation data only - never used for authorization.
	Payload   json.RawMessage `json:"payload"`
	ReadAt    *time.Time      `json:"read_at"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type TaskAssignedPayload struct {
	TaskID     uuid.UUID `json:"task_id"`
	TaskTitle  string    `json:"task_title"`
	ActorEmail string    `json:"actor_email"`
}

type WorkspaceInvitedPayload struct {
	InvitationID  uuid.UUID `json:"invitation_id"`
	WorkspaceID   uuid.UUID `json:"workspace_id"`
	WorkspaceName string    `json:"workspace_name"`
	ActorEmail    string    `json:"actor_email"`
}

type UnreadCountResponse struct {
	UnreadCount int `json:"unread_count"`
}
