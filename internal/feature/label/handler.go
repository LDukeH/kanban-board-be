package label

import (
	"encoding/json"
	"kanban-board-be/internal/apperrors"
	"kanban-board-be/internal/feature/auth"
	"log"
	"net/http"
	"regexp"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

var colorRegex = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$`)

func isValidColor(color string) bool {
	return color == "" || colorRegex.MatchString(color)
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) CreateLabel(w http.ResponseWriter, r *http.Request) {
	var req LabelCreateRequest

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
		http.Error(w, "Invalid label data", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Label name is required", http.StatusBadRequest)
		return
	}

	if !isValidColor(req.Color) {
		http.Error(w, "Invalid label color", http.StatusBadRequest)
		return
	}

	label := Label{
		BoardID: boardID,
		Name:    req.Name,
		Color:   req.Color,
	}

	if err := h.service.CreateLabel(r.Context(), userID, &label); err != nil {
		log.Printf("ERROR: Failed to create label: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(label)
}

func (h *Handler) GetLabelsByBoardID(w http.ResponseWriter, r *http.Request) {
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

	labels, err := h.service.GetLabelsByBoardID(r.Context(), userID, boardID)
	if err != nil {
		log.Printf("ERROR: Failed to get labels: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(labels)
}

func (h *Handler) GetLabelByID(w http.ResponseWriter, r *http.Request) {
	// Get label ID from URL
	labelIDStr := chi.URLParam(r, "labelID")
	labelID, err := uuid.Parse(labelIDStr)
	if err != nil {
		http.Error(w, "Invalid label ID", http.StatusBadRequest)
		return
	}

	// Get authenticated user ID
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	label, err := h.service.GetLabelByID(r.Context(), userID, labelID)
	if err != nil {
		log.Printf("ERROR: Failed to get label: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(label)
}

func (h *Handler) UpdateLabel(w http.ResponseWriter, r *http.Request) {
	var req LabelUpdateRequest
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	labelIDStr := chi.URLParam(r, "labelID")
	labelID, err := uuid.Parse(labelIDStr)
	if err != nil {
		http.Error(w, "Invalid label ID", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid label data", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Label name is required", http.StatusBadRequest)
		return
	}

	if !isValidColor(req.Color) {
		http.Error(w, "Invalid label color", http.StatusBadRequest)
		return
	}

	label := Label{
		ID:    labelID,
		Name:  req.Name,
		Color: req.Color,
	}

	if err := h.service.UpdateLabel(r.Context(), userID, &label); err != nil {
		log.Printf("ERROR: Failed to update label: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteLabel(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	labelIDStr := chi.URLParam(r, "labelID")
	labelID, err := uuid.Parse(labelIDStr)
	if err != nil {
		http.Error(w, "Invalid label ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteLabel(r.Context(), userID, labelID); err != nil {
		log.Printf("ERROR: Failed to delete label: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}