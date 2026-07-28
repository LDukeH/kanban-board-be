package board

import (
	"encoding/json"
	"log"
	"net/http"

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

func (h *Handler) CreateBoard(w http.ResponseWriter, r *http.Request) {
	var req BoardCreateRequest

	// Get workspace ID from URL
	workspaceIDStr := chi.URLParam(r, "workspaceID")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
		return
	}

	// Get authenticated user ID (for future authorization)
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
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
		Name: req.Name,
		WorkspaceID: workspaceID,
	}

	_ = userID

	if err := h.service.CreateBoard(r.Context(), uuid.MustParse(userID), &board); err != nil {
		log.Printf("ERROR: Failed to create board: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(board)
}