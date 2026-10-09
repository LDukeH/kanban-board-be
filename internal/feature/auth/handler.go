package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"kanban-board-be/internal/apperrors"
)

type Handler struct {
	authService *Service
}

func NewHandler(authService *Service) *Handler {
	return &Handler{authService: authService}
}

func (h *Handler) SyncUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}

	email, ok := r.Context().Value(UserEmailKey).(string)
	if !ok || email == "" {
		http.Error(w, "Email claim is required", http.StatusUnauthorized)
		return
	}

	if err := h.authService.SyncUser(r.Context(), userID, email); err != nil {
		log.Printf("Failed to sync Supabase user: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteMyAccountHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.authService.DeleteMyAccount(r.Context()); err != nil {
		log.Printf("Failed to delete account: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetMeHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.authService.LookupUser(r.Context())
	if err != nil {
		log.Printf("Failed to get user: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	getMeResponse := GetMeResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339Nano),
	}

	if err := json.NewEncoder(w).Encode(getMeResponse); err != nil {
		log.Printf("Failed to encode current user: %v", err)
	}
}
