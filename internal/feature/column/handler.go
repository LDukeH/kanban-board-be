package column

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

func (h *Handler) CreateColumn(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}

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

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid column data", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Column name is required", http.StatusBadRequest)
		return
	}

	column := Column{
		BoardID: boardID,
		Name:    req.Name,
	}

	if err := h.service.CreateColumn(r.Context(), userID, &column); err != nil {
		log.Printf("ERROR: Failed to create column: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(column)
}

func (h *Handler) GetColumnsByBoardID(w http.ResponseWriter, r *http.Request) {
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

	columns, err := h.service.GetColumnsByBoardID(r.Context(), boardID, userID)
	if err != nil {
		log.Printf("ERROR: Failed to get columns: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(columns)
}
