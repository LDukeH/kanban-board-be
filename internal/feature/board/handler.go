package board

import (
	"encoding/json"
	"kanban-board-be/internal/feature/auth"
	"log"
	"net/http"

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

func (h *Handler) CreateBoard(w http.ResponseWriter, r *http.Request) {
	var req BoardCreateRequest

	// Get workspace ID from URL
	workspaceIDStr := chi.URLParam(r, "workspaceID")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
		return
	}

	// Get authenticated user ID
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid board data", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Board name is required", http.StatusBadRequest)
		return
	}

	board := Board{
		Name:        req.Name,
		WorkspaceID: workspaceID,
	}

	if err := h.service.CreateBoard(r.Context(), userID, &board); err != nil {
		log.Printf("ERROR: Failed to create board: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(board)
}

func (h *Handler) GetBoardsByWorkspaceID(w http.ResponseWriter, r *http.Request) {
	// Get workspace ID from URL
	workspaceIDStr := chi.URLParam(r, "workspaceID")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
		return
	}

	// Get authenticated user ID
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	boards, err := h.service.GetBoardsByWorkspaceID(r.Context(), workspaceID, userID)
	if err != nil {
		log.Printf("ERROR: Failed to get boards: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(boards)
}

func (h *Handler) GetBoardByID(w http.ResponseWriter, r *http.Request) {
	// Get board ID from URL
	boardIDStr := chi.URLParam(r, "boardID")
	boardID, err := uuid.Parse(boardIDStr)
	if err != nil {
		http.Error(w, "Invalid board ID", http.StatusBadRequest)
		return
	}

	// Get authenticated user ID
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	board, err := h.service.GetBoardByID(r.Context(), boardID, userID)
	if err != nil {
		log.Printf("ERROR: Failed to get board: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(board)
}
