package invitation

import (
	"encoding/json"
	"log"
	"net/http"

	"kanban-board-be/internal/apperrors"
	"kanban-board-be/internal/feature/auth"
	"kanban-board-be/internal/feature/user"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	invitationService *Service
	userService       *user.Service
}

func NewHandler(invitationService *Service, userService *user.Service) *Handler {
	return &Handler{
		invitationService: invitationService,
		userService:       userService,
	}
}

func (h *Handler) workspaceInvitationResponse(r *http.Request, invitation *WorkspaceInvitation) *InvitationResponse {
	response := &InvitationResponse{
		ID:              invitation.ID,
		WorkspaceID:     invitation.WorkspaceID,
		InviteeUserID:   invitation.InviteeUserID,
		InvitedByUserID: invitation.InvitedByUserID,
		Status:          invitation.Status,
		CreatedAt:       invitation.CreatedAt,
		UpdatedAt:       invitation.UpdatedAt,
		ExpiresAt:       invitation.ExpiresAt,
	}

	if invitee, err := h.userService.GetUserByID(r.Context(), invitation.InviteeUserID.String()); err == nil {
		response.InviteeEmail = invitee.Email
	}

	return response
}

func (h *Handler) CreateInvitation(w http.ResponseWriter, r *http.Request) {
	workspaceIDStr := chi.URLParam(r, "workspaceID")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
		return
	}

	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req InvitationCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR decoding request body: %v\n", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == uuid.Nil {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	invitation, err := h.invitationService.Create(r.Context(), workspaceID, req.UserID, userID)
	if err != nil {
		log.Printf("ERROR creating invitation: %v\n", err)
		apperrors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(h.workspaceInvitationResponse(r, invitation))
}

func (h *Handler) GetInvitationsByWorkspaceID(w http.ResponseWriter, r *http.Request) {
	workspaceIDStr := chi.URLParam(r, "workspaceID")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
		return
	}

	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	invitations, err := h.invitationService.List(r.Context(), workspaceID, userID)
	if err != nil {
		log.Printf("ERROR listing invitations: %v\n", err)
		apperrors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invitations)
}

func (h *Handler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	invitationIDStr := chi.URLParam(r, "invitationID")
	invitationID, err := uuid.Parse(invitationIDStr)
	if err != nil {
		http.Error(w, "Invalid invitation ID", http.StatusBadRequest)
		return
	}

	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	invitation, err := h.invitationService.Accept(r.Context(), invitationID, userID)
	if err != nil {
		log.Printf("ERROR accepting invitation: %v\n", err)
		apperrors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(h.workspaceInvitationResponse(r, invitation))
}

func (h *Handler) DeclineInvitation(w http.ResponseWriter, r *http.Request) {
	invitationIDStr := chi.URLParam(r, "invitationID")
	invitationID, err := uuid.Parse(invitationIDStr)
	if err != nil {
		http.Error(w, "Invalid invitation ID", http.StatusBadRequest)
		return
	}

	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	invitation, err := h.invitationService.Decline(r.Context(), invitationID, userID)
	if err != nil {
		log.Printf("ERROR declining invitation: %v\n", err)
		apperrors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(h.workspaceInvitationResponse(r, invitation))
}

func (h *Handler) CancelInvitation(w http.ResponseWriter, r *http.Request) {
	workspaceIDStr := chi.URLParam(r, "workspaceID")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
		return
	}

	invitationIDStr := chi.URLParam(r, "invitationID")
	invitationID, err := uuid.Parse(invitationIDStr)
	if err != nil {
		http.Error(w, "Invalid invitation ID", http.StatusBadRequest)
		return
	}

	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	invitation, err := h.invitationService.Cancel(r.Context(), workspaceID, invitationID, userID)
	if err != nil {
		log.Printf("ERROR cancelling invitation: %v\n", err)
		apperrors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(h.workspaceInvitationResponse(r, invitation))
}
