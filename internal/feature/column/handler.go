package column

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

func (h *Handler) CreateColumn(w http.ResponseWriter, r *http.Request) {
	var req ColumnCreateRequest

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
		apperrors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(column)
}

func (h *Handler) GetColumnsByBoardID(w http.ResponseWriter, r *http.Request) {
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

	columns, err := h.service.GetColumnsByBoardID(r.Context(), boardID, userID)
	if err != nil {
		log.Printf("ERROR: Failed to get columns: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(columns)
}

func (h *Handler) UpdateColumn(w http.ResponseWriter, r *http.Request) {
	var req ColumnUpdateRequest

	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	columnIDStr := chi.URLParam(r, "columnID")
	columnID, err := uuid.Parse(columnIDStr)

	if err != nil {
		http.Error(w, "Invalid column ID", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid column data", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Column name is required", http.StatusBadRequest)
		return
	}

	column := Column{
		ID:   columnID,
		Name: req.Name,
	}

	if err := h.service.UpdateColumnName(r.Context(), userID, &column); err != nil {
		log.Printf("ERROR: Failed to update column: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UpdateColumnPosition(w http.ResponseWriter, r *http.Request) {
	var req ColumnPositionUpdateRequest

	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	columnIDStr := chi.URLParam(r, "columnID")
	columnID, err := uuid.Parse(columnIDStr)
	if err != nil {
		http.Error(w, "Invalid column ID", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	column := Column{
		ID:       columnID,
		Position: req.Position,
	}

	if err := h.service.UpdateColumnPosition(r.Context(), userID, &column); err != nil {
		log.Printf("ERROR: Failed to update column position: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteColumn(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	columnIDStr := chi.URLParam(r, "columnID")
	columnID, err := uuid.Parse(columnIDStr)
	if err != nil {
		http.Error(w, "Invalid column ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteColumn(r.Context(), userID, columnID); err != nil {
		log.Printf("ERROR: Failed to delete column: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
