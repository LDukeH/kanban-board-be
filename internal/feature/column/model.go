package column

import (
    "time"

    "github.com/google/uuid"
)

type Column struct {
    ID        uuid.UUID `json:"id"`
    BoardID   uuid.UUID `json:"board_id"`
    Name      string    `json:"name"`
    Position  int       `json:"position"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}