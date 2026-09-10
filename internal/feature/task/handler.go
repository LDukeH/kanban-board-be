package task

import (
	"encoding/json"
	"log"
	"net/http"

	"kanban-board-be/internal/apperrors"
	"kanban-board-be/internal/feature/auth"
	"kanban-board-be/internal/feature/label"
	"kanban-board-be/internal/feature/user"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	taskService  *Service
	labelService *label.Service
	userService  *user.Service
}

func NewHandler(service *Service, labelService *label.Service, userService *user.Service) *Handler {
	return &Handler{
		taskService:  service,
		labelService: labelService,
		userService:  userService,
	}
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req TaskCreateRequest

	columnIDStr := chi.URLParam(r, "columnID")
	columnID, err := uuid.Parse(columnIDStr)
	if err != nil {
		http.Error(w, "Invalid column ID", http.StatusBadRequest)
		return
	}

	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

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

	if err := h.taskService.CreateTask(r.Context(), userID, &task); err != nil {
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

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid task data", http.StatusBadRequest)
		return
	}

	task := Task{
		ID:          taskID,
		Title:       req.Title,
		Description: req.Description,
	}

	if err := h.taskService.UpdateTask(r.Context(), userID, &task); err != nil {
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

	if err := h.taskService.UpdateTaskPosition(r.Context(), userID, &task); err != nil {
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

	if err := h.taskService.DeleteTask(r.Context(), userID, taskID); err != nil {
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

	tasks, err := h.taskService.GetTasksByColumnID(r.Context(), columnID, userID)
	if err != nil {
		log.Printf("ERROR: Failed to get tasks by column ID: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
}

func (h *Handler) GetLabelsByTaskID(w http.ResponseWriter, r *http.Request) {
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

	labels, err := h.labelService.GetLabelsByTaskID(r.Context(), userID, taskID)
	if err != nil {
		log.Printf("ERROR: Failed to get labels for task: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(labels)
}

func (h *Handler) AddLabelToTask(w http.ResponseWriter, r *http.Request) {
	var req TaskLabelCreateRequest

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

	if err := h.labelService.AddLabelToTask(r.Context(), userID, taskID, req.LabelID); err != nil {
		log.Printf("ERROR: Failed to add label to task: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) RemoveLabelFromTask(w http.ResponseWriter, r *http.Request) {
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

	labelIDStr := chi.URLParam(r, "labelID")
	labelID, err := uuid.Parse(labelIDStr)
	if err != nil {
		http.Error(w, "Invalid label ID", http.StatusBadRequest)
		return
	}

	if err := h.labelService.RemoveLabelFromTask(r.Context(), userID, taskID, labelID); err != nil {
		log.Printf("ERROR: Failed to remove label from task: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AssignUserToTask(w http.ResponseWriter, r *http.Request) {
	var req TaskAssignUserRequest

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

	if err != nil {
		log.Printf("ERROR: Failed to find user by email: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	if err := h.taskService.AssignUserToTask(r.Context(), taskID, userID, req.UserID); err != nil {
		log.Printf("ERROR: Failed to assign user to task: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RemoveUserFromTask(w http.ResponseWriter, r *http.Request) {
	var req TaskRemoveUserRequest
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

	if err != nil {
		log.Printf("ERROR: Failed to find user by email: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	if err := h.taskService.RemoveUserFromTask(r.Context(), taskID, userID, req.UserID); err != nil {
		log.Printf("ERROR: Failed to remove user from task: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetUsersByTaskID(w http.ResponseWriter, r *http.Request) {
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

	users, err := h.taskService.GetUsersByTaskID(r.Context(), taskID, userID)
	if err != nil {
		log.Printf("ERROR: Failed to get users by task ID: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}
