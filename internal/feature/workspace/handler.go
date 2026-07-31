package workspace

import (
	"encoding/json"
	"log"
	"net/http"

	"kanban-board-be/internal/apperrors"
	"kanban-board-be/internal/feature/auth"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetAllWorkspaces(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	workspaces, err := h.service.GetAllWorkspaces(r.Context(), userID)
	if err != nil {
		log.Printf("ERROR fetching workspaces: %v\n", err)
		apperrors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workspaces)
}

func (h *Handler) GetWorkspaceByID(w http.ResponseWriter, r *http.Request) {
	// Extract the workspace ID from the URL path
	id := chi.URLParam(r, "workspaceID")

	// Check if the ID is a valid UUID
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	workspace, err := h.service.GetWorkspaceByID(r.Context(), id, userID)
	if err != nil {
		log.Printf("ERROR fetching workspace: %v\n", err)
		apperrors.WriteError(w, err)
		return
	}

	json.NewEncoder(w).Encode(workspace)
}

func (h *Handler) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	var req WorkspaceCreateRequest

	userID, err := auth.GetUserID(r)
	if err != nil {
		log.Println("ERROR: userID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR decoding request body: %v\n", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Workspace name is required", http.StatusBadRequest)
		return
	}

	workspace := Workspace{
		Name: req.Name,
	}

	if err := h.service.CreateWorkspace(r.Context(), &workspace, userID); err != nil {
		log.Printf("ERROR creating workspace: %v\n", err)
		apperrors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(workspace)
}

func (h *Handler) UpdateWorkspace(w http.ResponseWriter, r *http.Request) {
	var req WorkspaceUpdateRequest
	userID, err := auth.GetUserID(r)
	if err != nil {
		log.Println("ERROR: userID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	workspaceIDStr := chi.URLParam(r, "workspaceID")
	workspaceID, err := uuid.Parse(workspaceIDStr)

	if err != nil {
		log.Println("ERROR: userID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR decoding request body: %v\n", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Workspace name is required", http.StatusBadRequest)
		return
	}

	workspace := Workspace{
		Name: req.Name,
		ID:   workspaceID,
	}

	if err := h.service.UpdateWorkspace(r.Context(), workspace, userID); err != nil {
		log.Printf("ERROR updating workspace: %v\n", apperrors.ErrConflict)
		apperrors.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	// Extract the workspace ID from the URL path
	workspaceIDStr := chi.URLParam(r, "workspaceID")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
		return
	}

	userID, err := auth.GetUserID(r)
	if err != nil {
		log.Println("ERROR: userID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.service.DeleteWorkspace(r.Context(), workspaceID, userID); err != nil {
		log.Printf("ERROR deleting workspace: %v\n", err)
		apperrors.WriteError(w, err)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}