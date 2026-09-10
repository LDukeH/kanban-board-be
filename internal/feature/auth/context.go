package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

// GetUserID safely extracts and parses the authenticated user ID from the request context.
// Returns an error if the user ID is missing, not a string, or not a valid UUID.
func GetUserID(r *http.Request) (uuid.UUID, error) {
	userIDStr, ok := r.Context().Value(UserIDKey).(string)
	if !ok || userIDStr == "" {
		return uuid.Nil, fmt.Errorf("unauthorized")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID: %w", err)
	}

	return userID, nil
}

// GetUserIDCtx is a context-based variant for non-http code paths.
// Kept for consistency; most callers should use GetUserID.
func GetUserIDCtx(ctx context.Context) (uuid.UUID, error) {
	userIDStr, ok := ctx.Value(UserIDKey).(string)
	if !ok || userIDStr == "" {
		return uuid.Nil, fmt.Errorf("unauthorized")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID: %w", err)
	}

	return userID, nil
}
