package invitation

import (
	"time"

	"github.com/google/uuid"
)

// Status is the lifecycle state of a workspace invitation.
type Status string

const (
	StatusPending   Status = "pending"
	StatusAccepted  Status = "accepted"
	StatusDeclined  Status = "declined"
	StatusCancelled Status = "cancelled"
	StatusExpired   Status = "expired"
)

// DefaultExpiry is how long a pending invitation stays acceptable.
const DefaultExpiry = 7 * 24 * time.Hour

type WorkspaceInvitation struct {
	ID              uuid.UUID `json:"id"`
	WorkspaceID     uuid.UUID `json:"workspace_id"`
	InviteeUserID   uuid.UUID `json:"invitee_user_id"`
	InvitedByUserID uuid.UUID `json:"invited_by_user_id"`
	Status          Status    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	ExpiresAt       time.Time `json:"expires_at"`
}

// EffectiveStatus reports the invitation's current status, treating a pending
// invitation whose expiry has passed as expired.
func (i *WorkspaceInvitation) EffectiveStatus(now time.Time) Status {
	return effectiveStatus(i.Status, i.ExpiresAt, now)
}

func effectiveStatus(status Status, expiresAt, now time.Time) Status {
	if status == StatusPending && now.After(expiresAt) {
		return StatusExpired
	}
	return status
}

type InvitationCreateRequest struct {
	// UserID identifies the registered user to invite.
	UserID uuid.UUID `json:"user_id"`
}

type InvitationResponse struct {
	ID              uuid.UUID `json:"id"`
	WorkspaceID     uuid.UUID `json:"workspace_id"`
	InviteeUserID   uuid.UUID `json:"invitee_user_id"`
	InvitedByUserID uuid.UUID `json:"invited_by_user_id"`
	Status          Status    `json:"status"`
	InviteeEmail    string    `json:"invitee_email"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	ExpiresAt       time.Time `json:"expires_at"`
}
