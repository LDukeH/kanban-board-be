package label

import (
	"time"

	"github.com/google/uuid"
)

type Label struct {
    ID        uuid.UUID `json:"id"`
    BoardID   uuid.UUID `json:"board_id"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}