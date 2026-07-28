package notification

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
    ID        uuid.UUID `json:"id"`
    UserID    uuid.UUID `json:"user_id"`
    Type      string    `json:"type"`
    Message   string    `json:"message"`
    IsRead    bool      `json:"is_read"`
    CreatedAt time.Time `json:"created_at"`
}