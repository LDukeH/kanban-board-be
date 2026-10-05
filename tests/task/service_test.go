package task_test

import (
	"context"
	"errors"
	"testing"

	"kanban-board-be/internal/feature/membership"
	. "kanban-board-be/internal/feature/task"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type fakeTaskRepo struct {
	created   []*Task
	updated   []*Task
	deleted   []uuid.UUID
	createErr error
	updateErr error
	deleteErr error
	byID      map[uuid.UUID]*Task
	byIDErr   error
}

func (f *fakeTaskRepo) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return nil, errors.New("BeginTx not implemented in fake")
}

func (f *fakeTaskRepo) Create(ctx context.Context, task *Task) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created = append(f.created, task)
	return nil
}

func (f *fakeTaskRepo) UpdateTask(ctx context.Context, task *Task) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.updated = append(f.updated, task)
	return nil
}

func (f *fakeTaskRepo) GetByID(ctx context.Context, taskID uuid.UUID) (*Task, error) {
	if f.byIDErr != nil {
		return nil, f.byIDErr
	}
	if f.byID != nil {
		if task, ok := f.byID[taskID]; ok {
			return task, nil
		}
	}
	return &Task{ID: taskID}, nil
}

func (f *fakeTaskRepo) UpdateTaskPosition(ctx context.Context, task *Task) error {
	return nil
}

func (f *fakeTaskRepo) DeleteTask(ctx context.Context, taskID uuid.UUID, version int) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	f.deleted = append(f.deleted, taskID)
	return nil
}

func (f *fakeTaskRepo) GetByColumnID(ctx context.Context, columnID uuid.UUID) ([]*Task, error) {
	return nil, nil
}

func (f *fakeTaskRepo) AssignUserToTaskTx(ctx context.Context, tx pgx.Tx, taskID uuid.UUID, userID uuid.UUID) error {
	return nil
}

func (f *fakeTaskRepo) RemoveUserFromTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error {
	return nil
}

func (f *fakeTaskRepo) IsUserAssignedToTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) (bool, error) {
	return false, nil
}

func (f *fakeTaskRepo) GetUsersByTaskID(ctx context.Context, taskID uuid.UUID) ([]*UsersTaskResponse, error) {
	return nil, nil
}

type stubResolver struct {
	role membership.WorkspaceRole
	err  error
}

func (s stubResolver) GetWorkspaceRole(ctx context.Context, userID, workspaceID uuid.UUID) (membership.WorkspaceRole, error) {
	return s.role, s.err
}

func (s stubResolver) GetBoardRole(ctx context.Context, userID, boardID uuid.UUID) (uuid.UUID, membership.WorkspaceRole, error) {
	return uuid.Nil, s.role, s.err
}

func (s stubResolver) GetColumnRole(ctx context.Context, userID, columnID uuid.UUID) (uuid.UUID, membership.WorkspaceRole, error) {
	return uuid.Nil, s.role, s.err
}

func (s stubResolver) GetTaskRole(ctx context.Context, userID, taskID uuid.UUID) (uuid.UUID, uuid.UUID, membership.WorkspaceRole, error) {
	return uuid.Nil, uuid.Nil, s.role, s.err
}

func (s stubResolver) GetLabelRole(ctx context.Context, userID, labelID uuid.UUID) (uuid.UUID, membership.WorkspaceRole, error) {
	return uuid.Nil, s.role, s.err
}

func serviceWithRole(t *testing.T, repo TaskRepository, role membership.WorkspaceRole, err error) *Service {
	t.Helper()
	authzSvc := newAuthzServiceForTest(t, stubResolver{role: role, err: err})
	return NewService(repo, authzSvc, nil, nil)
}
