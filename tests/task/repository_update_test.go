package task_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"kanban-board-be/internal/apperrors"
	. "kanban-board-be/internal/feature/task"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
)

// mockRepo builds a repository over a pgxmock pool.
func mockRepo(t *testing.T) (*Repository, pgxmock.PgxPoolIface) {
	t.Helper()
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool: %v", err)
	}
	return NewRepositoryWithDB(mock), mock
}

// The two halves of the guard: the row must still exist, and its version must
// match. Both are asserted separately so a regression cannot drop one silently.
var (
	guardExistence = regexp.MustCompile(`SELECT EXISTS \(SELECT 1 FROM tasks WHERE id = \$1\)`)
)

func TestRepositoryClassifyVersionMismatch(t *testing.T) {
	if err := ClassifyVersionMismatch(true); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("exists -> %v, want ErrConflict", err)
	}
	if err := ClassifyVersionMismatch(false); !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("missing -> %v, want ErrNotFound", err)
	}
}

func TestRepositoryUpdateTaskSuccess(t *testing.T) {
	repo, mock := mockRepo(t)
	taskID := uuid.New()

	mock.ExpectExec(`UPDATE tasks`).
		WithArgs("renamed", taskID, 1).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	task := &Task{ID: taskID, Title: strp("renamed"), Version: 1}
	if err := repo.UpdateTask(context.Background(), task); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	if task.Version != 2 {
		t.Fatalf("version = %d, want 2 (bumped for the caller)", task.Version)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// Two clients read version 1. The first write wins; the second is refused.
func TestRepositoryUpdateTaskTwoClientsSameVersion(t *testing.T) {
	repo, mock := mockRepo(t)
	taskID := uuid.New()

	mock.ExpectExec(`UPDATE tasks`).
		WithArgs("first", taskID, 1).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	mock.ExpectExec(`UPDATE tasks`).
		WithArgs("second", taskID, 1).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))
	mock.ExpectQuery(guardExistence.String()).
		WithArgs(taskID).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

	first := &Task{ID: taskID, Title: strp("first"), Version: 1}
	if err := repo.UpdateTask(context.Background(), first); err != nil {
		t.Fatalf("first UpdateTask: %v", err)
	}
	if first.Version != 2 {
		t.Fatalf("first version = %d, want 2", first.Version)
	}

	second := &Task{ID: taskID, Title: strp("second"), Version: 1}
	if err := repo.UpdateTask(context.Background(), second); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("second UpdateTask = %v, want ErrConflict", err)
	}
	if second.Version != 1 {
		t.Fatalf("second version = %d, want it left at 1", second.Version)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryUpdateTaskMissingRowIsNotFound(t *testing.T) {
	repo, mock := mockRepo(t)
	taskID := uuid.New()

	mock.ExpectExec(`UPDATE tasks`).
		WithArgs("t", taskID, 1).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))
	mock.ExpectQuery(guardExistence.String()).
		WithArgs(taskID).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(false))

	err := repo.UpdateTask(context.Background(), &Task{ID: taskID, Title: strp("t"), Version: 1})
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryUpdateTaskEmptyPatchStillGuards(t *testing.T) {
	repo, mock := mockRepo(t)
	taskID := uuid.New()

	mock.ExpectQuery(`SELECT version FROM tasks WHERE id = \$1`).
		WithArgs(taskID).
		WillReturnRows(pgxmock.NewRows([]string{"version"}).AddRow(4))

	task := &Task{ID: taskID, Version: 4}
	if err := repo.UpdateTask(context.Background(), task); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	if task.Version != 4 {
		t.Fatalf("version = %d, want 4 (no write, so no bump)", task.Version)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryUpdateTaskEmptyPatchStaleIsConflict(t *testing.T) {
	repo, mock := mockRepo(t)
	taskID := uuid.New()

	mock.ExpectQuery(`SELECT version FROM tasks WHERE id = \$1`).
		WithArgs(taskID).
		WillReturnRows(pgxmock.NewRows([]string{"version"}).AddRow(5))

	err := repo.UpdateTask(context.Background(), &Task{ID: taskID, Version: 4})
	if !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
