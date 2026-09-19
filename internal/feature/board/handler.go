package board

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

func (h *Handler) CreateBoard(w http.ResponseWriter, r *http.Request) {
	var req BoardCreateRequest

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
		apperrors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(board)
}

func (h *Handler) GetBoardsByWorkspaceID(w http.ResponseWriter, r *http.Request) {
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

	boards, err := h.service.GetBoardsByWorkspaceID(r.Context(), workspaceID, userID)
	if err != nil {
		log.Printf("ERROR: Failed to get boards: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(boards)
}

func (h *Handler) GetBoardByID(w http.ResponseWriter, r *http.Request) {
	boardIDStr := chi.URLParam(r, "boardID")
	boardID, err := uuid.Parse(boardIDStr)
	if err != nil {
		http.Error(w, "Invalid board ID", http.StatusBadRequest)
		return
	}

	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	board, err := h.service.GetBoardByID(r.Context(), boardID, userID)
	if err != nil {
		log.Printf("ERROR: Failed to get board: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(board)
}

func (h *Handler) UpdateBoard(w http.ResponseWriter, r *http.Request) {
	var req BoardUpdateRequest

	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	boardIDStr := chi.URLParam(r, "boardID")
	boardID, err := uuid.Parse(boardIDStr)

	if err != nil {
		log.Println("ERROR: Invalid board ID")
		http.Error(w, "Invalid board ID", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR decoding request body: %v\n", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Board name is required", http.StatusBadRequest)
		return
	}

	board := Board{
		Name: req.Name,
		ID:   boardID,
	}

	if err := h.service.UpdateBoard(r.Context(), userID, &board); err != nil {
		log.Printf("ERROR: Failed to update board: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	// Return the updated board so clients can read the new name/timestamps.
	updated, err := h.service.GetBoardByID(r.Context(), boardID, userID)
	if err != nil {
		log.Printf("ERROR: Failed to read updated board: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}

func (h *Handler) DeleteBoard(w http.ResponseWriter, r *http.Request) {
	boardIDStr := chi.URLParam(r, "boardID")

	boardID, err := uuid.Parse(boardIDStr)
	if err != nil {
		http.Error(w, "Invalid board ID", http.StatusBadRequest)
		return
	}

	userID, err := auth.GetUserID(r)
	if err != nil {
		log.Println("ERROR: userID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.service.DeleteBoard(r.Context(), boardID, userID); err != nil {
		log.Printf("ERROR: Failed to delete board: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
