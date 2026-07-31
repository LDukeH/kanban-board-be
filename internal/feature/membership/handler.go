package membership

import (
	"encoding/json"
	"fmt"
	"kanban-board-be/internal/feature/auth"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	membershipService *Service
}

func NewHandler(membershipService *Service) *Handler {
	return &Handler{
		membershipService: membershipService,
	}
}

func (h* Handler) CreateMembership(w http.ResponseWriter, r *http.Request) {
	var req MembershipCreateRequest
	workspaceIDStr := chi.URLParam(r, "workspaceID")
	workspaceID, err := uuid.Parse(workspaceIDStr)

	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR decoding request body: %v\n", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	fmt.Printf("WorkspaceID: %s, UserID: %s, Role: %s\n", workspaceID, req.UserID, req.Role)

	if( workspaceID == uuid.Nil || req.UserID == uuid.Nil || req.Role == "" ) {
		http.Error(w, "WorkspaceID, UserID and Role are required", http.StatusBadRequest)
		return
	}

	workspaceMembership := &WorkspaceMembership{
		WorkspaceID: workspaceID,
		UserID:      req.UserID,
		Role:        req.Role,
	}

	if err := h.membershipService.CreateMembership(r.Context(), workspaceMembership, userID); err != nil {
		log.Printf("ERROR creating membership: %v\n", err)
		http.Error(w, "Failed to create membership", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(workspaceMembership)

}

func (h *Handler) ChangeMemberRole(w http.ResponseWriter, r *http.Request) {
	var req ChangeRoleRequest
	membershipIDStr := chi.URLParam(r, "id")
    membershipID, err := uuid.Parse(membershipIDStr)
    if err != nil {
        http.Error(w, "Invalid membership ID", http.StatusBadRequest)
        return
    }

    workspaceIDStr := chi.URLParam(r, "workspaceID")
    workspaceID, err := uuid.Parse(workspaceIDStr)
    if err != nil {
        http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
        return
    }

	// The one who request
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}


	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR decoding request body: %v\n", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Role == "" {
		http.Error(w, "Role are required", http.StatusBadRequest)
		return
	}
	
	workspaceMembership := &WorkspaceMembership{
		ID: membershipID,
		WorkspaceID: workspaceID,
		Role: req.Role,
	}

	if err := h.membershipService.ChangeMemberRole(r.Context(), workspaceMembership, userID); err != nil {
		log.Printf("ERROR changing member role: %v\n", err)
		http.Error(w, "Failed to change member role", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	membershipIDStr := chi.URLParam(r, "id")
    membershipID, err := uuid.Parse(membershipIDStr)
    if err != nil {
        http.Error(w, "Invalid membership ID", http.StatusBadRequest)
        return
    }

    workspaceIDStr := chi.URLParam(r, "workspaceID")
    workspaceID, err := uuid.Parse(workspaceIDStr)
    if err != nil {
        http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
        return
    }

	// The one who request
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}


	workspaceMembership := &WorkspaceMembership{
		ID: membershipID,
		WorkspaceID: workspaceID,
	}

	if err := h.membershipService.RemoveMember(r.Context(), workspaceMembership, userID); err != nil {
		log.Printf("ERROR removing member: %v\n", err)
		http.Error(w, "Failed to remove member", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetMembershipsByWorkspaceID(w http.ResponseWriter, r *http.Request) {
	workspaceIDStr := chi.URLParam(r, "workspaceID")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		log.Printf("ERROR parsing workspace ID: %v\n", err)
		http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
		return
	}

	userID, err := auth.GetUserID(r)
	if err != nil {
		log.Printf("ERROR parsing user ID: %v\n", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	memberships, err := h.membershipService.GetMemberListByWorkspaceID(r.Context(), workspaceID, userID)
	if err != nil {
		log.Printf("ERROR getting memberships: %v\n", err)
		http.Error(w, "Failed to get memberships", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(memberships)
}