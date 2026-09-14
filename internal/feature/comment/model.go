package comment

import (
	"time"

	"github.com/google/uuid"
)

// Comment is the database-backed model for a comment row. A comment is
// immutable except deletion: there is no update endpoint.
type Comment struct {
	ID        uuid.UUID `json:"id"`
	TaskID    uuid.UUID `json:"task_id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CommentResponse is the API shape returned to the frontend. author_email is
// populated by joining users in the repository; it is not a database column.
type CommentResponse struct {
	ID          uuid.UUID `json:"id"`
	TaskID      uuid.UUID `json:"task_id"`
	AuthorID    uuid.UUID `json:"author_id"`
	AuthorEmail string    `json:"author_email"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateCommentRequest is the body for POST /tasks/{taskID}/comments.
// The task ID comes from the URL and the author ID comes from the
// authenticated session; neither is accepted from the client.
type CreateCommentRequest struct {
	Content string `json:"content"`
}
