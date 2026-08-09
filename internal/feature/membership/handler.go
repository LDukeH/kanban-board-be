package membership

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"kanban-board-be/internal/apperrors"
	"kanban-board-be/internal/feature/auth"
	"kanban-board-be/internal/feature/user"
)

type Handler struct {
	membershipService *Service
	userService *user.Service
}

func NewHandler(membershipService *Service, userService *user.Service) *Handler {
	return &Handler{
		membershipService: membershipService,
		userService: userService,
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

	if( workspaceID == uuid.Nil || req.Email == "" || req.Role == "" ) {
		http.Error(w, "WorkspaceID, Email and Role are required", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		log.Printf("ERROR finding user: %v\n", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	workspaceMembership := &WorkspaceMembership{
		WorkspaceID: workspaceID,
		UserID:      user.ID,
		Role:        req.Role,
	}

	if err := h.membershipService.CreateMembership(r.Context(), workspaceMembership, userID); err != nil {
		log.Printf("ERROR creating membership: %v\n", err)
		apperrors.WriteError(w, err)
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
		apperrors.WriteError(w, err)
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
		apperrors.WriteError(w, err)
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
		apperrors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(memberships)
}