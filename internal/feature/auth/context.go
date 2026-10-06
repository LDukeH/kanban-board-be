package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

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
