package task

import (
	"encoding/json"
	"kanban-board-be/internal/apperrors"
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

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req TaskCreateRequest

	// Get column ID from URL
	columnIDStr := chi.URLParam(r, "columnID")
	columnID, err := uuid.Parse(columnIDStr)
	if err != nil {
		http.Error(w, "Invalid column ID", http.StatusBadRequest)
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
		http.Error(w, "Invalid task data", http.StatusBadRequest)
		return
	}

	if req.Title == nil || *req.Title == "" {
		http.Error(w, "Task title is required", http.StatusBadRequest)
		return
	}

	task := Task{
		ColumnID:    columnID,
		Title:       req.Title,
		Description: req.Description,
	}

	if err := h.service.CreateTask(r.Context(), userID, &task); err != nil {
		log.Printf("ERROR: Failed to create task: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	var req TaskUpdateRequest

	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	taskIDStr := chi.URLParam(r, "taskID")
	taskID, err := uuid.Parse(taskIDStr)
	
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid task data", http.StatusBadRequest)
		return
	}

	task := Task{
		ID:          taskID,
		Title:       req.Title,
		Description: req.Description,
	}

	if err := h.service.UpdateTask(r.Context(), userID, &task); err != nil {
		log.Printf("ERROR: Failed to update task: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UpdateTaskPosition(w http.ResponseWriter, r *http.Request) {
	var req TaskPositionUpdateRequest

	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	taskIDStr := chi.URLParam(r, "taskID")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	task := Task{
		ID:       taskID,
		ColumnID: req.ColumnID,
		Position: req.Position,
	}


	if err := h.service.UpdateTaskPosition(r.Context(), userID, &task); err != nil {
		log.Printf("ERROR: Failed to update task position: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	taskIDStr := chi.URLParam(r, "taskID")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteTask(r.Context(), userID, taskID); err != nil {
		log.Printf("ERROR: Failed to delete task: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetTasksByColumnID(w http.ResponseWriter, r *http.Request) {
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

	tasks, err := h.service.GetTasksByColumnID(r.Context(), columnID, userID)
	if err != nil {
		log.Printf("ERROR: Failed to get tasks by column ID: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
}