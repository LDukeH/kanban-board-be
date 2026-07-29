package workspace

import (
	"encoding/json"
	"kanban-board-be/internal/feature/auth"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workspaces)
}

func (h *Handler) GetWorkspaceByID(w http.ResponseWriter, r *http.Request) {
	// Extract the workspace ID from the URL path
	id := chi.URLParam(r, "id")

	//TODO: missing checking for user role

	workspace, err := h.service.GetWorkspaceByID(r.Context(), id)
	if err != nil {
		log.Printf("ERROR fetching workspace: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(workspace)
}
