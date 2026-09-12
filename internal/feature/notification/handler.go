package notification

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"kanban-board-be/internal/apperrors"
	"kanban-board-be/internal/feature/auth"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	notificationService *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		notificationService: service,
	}
}

func (h *Handler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// limit defaults to 20 and must be in the range 1..50.
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil || parsed < 1 || parsed > 50 {
			http.Error(w, "Invalid query parameters", http.StatusBadRequest)
			return
		}
		limit = parsed
	}

	unreadOnly := false
	if unreadStr := r.URL.Query().Get("unread_only"); unreadStr != "" {
		parsed, err := strconv.ParseBool(unreadStr)
		if err != nil {
			http.Error(w, "Invalid query parameters", http.StatusBadRequest)
			return
		}
		unreadOnly = parsed
	}

	notifications, err := h.notificationService.List(r.Context(), userID, limit, unreadOnly)
	if err != nil {
		log.Printf("ERROR: Failed to list notifications: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notifications)
}

func (h *Handler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	count, err := h.notificationService.UnreadCount(r.Context(), userID)
	if err != nil {
		log.Printf("ERROR: Failed to get unread notification count: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(UnreadCountResponse{UnreadCount: count})
}

func (h *Handler) MarkNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	notificationIDStr := chi.URLParam(r, "notificationID")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	if err := h.notificationService.MarkAsRead(r.Context(), notificationID, userID); err != nil {
		log.Printf("ERROR: Failed to mark notification as read: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) MarkAllNotificationsAsRead(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.notificationService.MarkAllAsRead(r.Context(), userID); err != nil {
		log.Printf("ERROR: Failed to mark all notifications as read: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
