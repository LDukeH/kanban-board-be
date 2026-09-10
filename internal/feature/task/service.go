package task

import (
	"context"
	"fmt"
	"kanban-board-be/internal/apperrors"
	"kanban-board-be/internal/feature/authz"
	"kanban-board-be/internal/feature/membership"

	"github.com/google/uuid"
)

type taskRepository interface {
	Create(ctx context.Context, task *Task) error
	UpdateTask(ctx context.Context, task *Task) error
	GetByID(ctx context.Context, taskID uuid.UUID) (*Task, error)
	UpdateTaskPosition(ctx context.Context, task *Task) error
	DeleteTask(ctx context.Context, taskID uuid.UUID) error
	GetByColumnID(ctx context.Context, columnID uuid.UUID) ([]*Task, error)
	AssignUserToTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error
	RemoveUserFromTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error
	IsUserAssignedToTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) (bool, error)
	GetUsersByTaskID(ctx context.Context, taskID uuid.UUID) ([]*UsersTaskResponse, error)
}

type Service struct {
	taskRepo taskRepository
	authz    *authz.Service
}

func NewService(taskRepo taskRepository, authz *authz.Service) *Service {
	return &Service{
		taskRepo: taskRepo,
		authz:    authz,
	}
}

// all member can freely work on tasks, so we only check if the user is a member of the workspace
func (s *Service) CreateTask(ctx context.Context, userID uuid.UUID, task *Task) error {
	role, err := s.authz.GetColumnRole(ctx, userID, task.ColumnID)
	if err != nil {
		return err
	}

	if role != membership.Owner && role != membership.Admin && role != membership.Member {
		return apperrors.ErrForbidden
	}
	return s.taskRepo.Create(ctx, task)
}

func (s *Service) UpdateTask(ctx context.Context, userID uuid.UUID, task *Task) error {
	role, err := s.authz.GetTaskRole(ctx, userID, task.ID)
	if err != nil {
		return err
	}

	if role != membership.Owner && role != membership.Admin && role != membership.Member {
		return apperrors.ErrForbidden
	}
	return s.taskRepo.UpdateTask(ctx, task)
}

func (s *Service) UpdateTaskPosition(ctx context.Context, userID uuid.UUID, task *Task) error {
	// Fetch the current task to get the source column.
	currentTask, err := s.taskRepo.GetByID(ctx, task.ID)
	if err != nil {
		return err
	}

	// Authorize based on the task's current column.
	role, err := s.authz.GetColumnRole(ctx, userID, currentTask.ColumnID)
	if err != nil {
		return err
	}

	if role != membership.Owner && role != membership.Admin && role != membership.Member {
		return apperrors.ErrForbidden
	}

	// Nothing to validate if the task stays in the same column.
	if task.ColumnID == currentTask.ColumnID {
		return s.taskRepo.UpdateTaskPosition(ctx, task)
	}

	// Target column must exist and belong to the same board.
	sourceBoardID, err := s.authz.GetColumnBoardID(ctx, currentTask.ColumnID)
	if err != nil {
		return err
	}

	targetBoardID, err := s.authz.GetColumnBoardID(ctx, task.ColumnID)
	if err != nil {
		return err
	}

	if sourceBoardID != targetBoardID {
		return apperrors.ErrForbidden
	}

	return s.taskRepo.UpdateTaskPosition(ctx, task)
}

func (s *Service) DeleteTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) error {
	role, err := s.authz.GetTaskRole(ctx, userID, taskID)
	if err != nil {
		return err
	}

	if role != membership.Owner && role != membership.Admin {
		return apperrors.ErrForbidden
	}

	return s.taskRepo.DeleteTask(ctx, taskID)
}

func (s *Service) GetTasksByColumnID(ctx context.Context, columnID uuid.UUID, userID uuid.UUID) ([]*Task, error) {
	// Check if the user has access to the column's board and workspace
	if _, err := s.authz.GetColumnRole(ctx, userID, columnID); err != nil {
		return nil, err
	}

	return s.taskRepo.GetByColumnID(ctx, columnID)
}

func (s *Service) AssignUserToTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, assignedUserID uuid.UUID) error {
	role, err := s.authz.GetTaskRole(ctx, userID, taskID)

	if err != nil {
		return err
	}

	if role != membership.Owner && role != membership.Admin && role != membership.Member {
		return apperrors.ErrForbidden
	}

	roleAssignedUser, err := s.authz.GetTaskRole(ctx, assignedUserID, taskID)
	if err != nil {
		return err
	}

	if roleAssignedUser != membership.Owner && roleAssignedUser != membership.Admin && roleAssignedUser != membership.Member {
		fmt.Println("Assigned user is not a member of the workspace")
		return apperrors.ErrForbidden
	}

	return s.taskRepo.AssignUserToTask(ctx, taskID, assignedUserID)
}

func (s *Service) RemoveUserFromTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, removedUserID uuid.UUID) error {
	role, err := s.authz.GetTaskRole(ctx, userID, taskID)

	if err != nil {
		return err
	}

	isSelfRemoval := userID == removedUserID

	if !isSelfRemoval && role != membership.Owner && role != membership.Admin {
		return apperrors.ErrForbidden
	}

	isAssigned, err := s.taskRepo.IsUserAssignedToTask(ctx, taskID, removedUserID)
	if err != nil {
		return err
	}
	if !isAssigned {
		fmt.Println("User is not assigned to this task")
		return apperrors.ErrNotFound
	}

	return s.taskRepo.RemoveUserFromTask(ctx, taskID, removedUserID)
}

func (s *Service) GetUsersByTaskID(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) ([]*UsersTaskResponse, error) {
	// Check if the user has access to the task's column, board, and workspace
	if _, err := s.authz.GetTaskRole(ctx, userID, taskID); err != nil {
		return nil, err
	}
	return s.taskRepo.GetUsersByTaskID(ctx, taskID)
}
