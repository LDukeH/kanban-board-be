package task_test

import (
	"context"
	"errors"
	"testing"

	"kanban-board-be/internal/apperrors"
	. "kanban-board-be/internal/feature/task"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
)

// expectCurrentTask queues the lookup that opens a move or delete.
func expectCurrentTask(mock pgxmock.PgxPoolIface, taskID, columnID uuid.UUID, position int) {
	mock.ExpectQuery(`SELECT id, column_id, position FROM tasks`).
		WithArgs(taskID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "column_id", "position"}).AddRow(taskID, columnID, position))
}

func TestRepositoryMoveSameColumnSuccess(t *testing.T) {
	repo, mock := mockRepo(t)
	taskID, columnID := uuid.New(), uuid.New()

	mock.ExpectBegin()
	expectCurrentTask(mock, taskID, columnID, 1)
	// Downwards: the tasks between old and new move up one slot.
	mock.ExpectExec(`UPDATE tasks`).
		WithArgs(columnID, 1, 3).
		WillReturnResult(pgxmock.NewResult("UPDATE", 2))
	mock.ExpectExec(`UPDATE tasks`).
		WithArgs(3, columnID, taskID, 1).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectCommit()

	task := &Task{ID: taskID, ColumnID: columnID, Position: 3, Version: 1}
	if err := repo.UpdateTaskPosition(context.Background(), task); err != nil {
		t.Fatalf("UpdateTaskPosition: %v", err)
	}
	if task.Version != 2 {
		t.Fatalf("version = %d, want 2", task.Version)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryMoveCrossColumnSuccess(t *testing.T) {
	repo, mock := mockRepo(t)
	taskID, sourceID, targetID := uuid.New(), uuid.New(), uuid.New()

	mock.ExpectBegin()
	expectCurrentTask(mock, taskID, sourceID, 0)
	mock.ExpectExec(`UPDATE tasks`).
		WithArgs(sourceID, 0).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectExec(`UPDATE tasks`).
		WithArgs(targetID, 2).
		WillReturnResult(pgxmock.NewResult("UPDATE", 3))
	mock.ExpectExec(`UPDATE tasks`).
		WithArgs(2, targetID, taskID, 1).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectCommit()

	task := &Task{ID: taskID, ColumnID: targetID, Position: 2, Version: 1}
	if err := repo.UpdateTaskPosition(context.Background(), task); err != nil {
		t.Fatalf("UpdateTaskPosition: %v", err)
	}
	if task.Version != 2 {
		t.Fatalf("version = %d, want 2", task.Version)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// A stale destination must roll the whole reorder back: the shifts already ran,
// so nothing may commit. ExpectCommit is deliberately not queued, so any commit
// call fails the test as an unexpected call.
func TestRepositoryMoveStaleVersionRollsBack(t *testing.T) {
	repo, mock := mockRepo(t)
	taskID, columnID := uuid.New(), uuid.New()

	mock.ExpectBegin()
	expectCurrentTask(mock, taskID, columnID, 1)
	mock.ExpectExec(`UPDATE tasks`).
		WithArgs(columnID, 1, 3).
		WillReturnResult(pgxmock.NewResult("UPDATE", 2))
	mock.ExpectExec(`UPDATE tasks`).
		WithArgs(3, columnID, taskID, 1).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))
	mock.ExpectQuery(guardExistence.String()).
		WithArgs(taskID).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectRollback()

	task := &Task{ID: taskID, ColumnID: columnID, Position: 3, Version: 1}
	if err := repo.UpdateTaskPosition(context.Background(), task); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
	if task.Version != 1 {
		t.Fatalf("version = %d, want it left at 1", task.Version)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryMoveStaleVersionCrossColumnRollsBack(t *testing.T) {
	repo, mock := mockRepo(t)
	taskID, sourceID, targetID := uuid.New(), uuid.New(), uuid.New()

	mock.ExpectBegin()
	expectCurrentTask(mock, taskID, sourceID, 0)
	mock.ExpectExec(`UPDATE tasks`).
		WithArgs(sourceID, 0).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectExec(`UPDATE tasks`).
		WithArgs(targetID, 2).
		WillReturnResult(pgxmock.NewResult("UPDATE", 3))
	mock.ExpectExec(`UPDATE tasks`).
		WithArgs(2, targetID, taskID, 1).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))
	mock.ExpectQuery(guardExistence.String()).
		WithArgs(taskID).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectRollback()

	task := &Task{ID: taskID, ColumnID: targetID, Position: 2, Version: 1}
	if err := repo.UpdateTaskPosition(context.Background(), task); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryDeleteTaskSuccess(t *testing.T) {
	repo, mock := mockRepo(t)
	taskID, columnID := uuid.New(), uuid.New()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT column_id, position FROM tasks`).
		WithArgs(taskID).
		WillReturnRows(pgxmock.NewRows([]string{"column_id", "position"}).AddRow(columnID, 2))
	mock.ExpectExec(`DELETE FROM tasks`).
		WithArgs(taskID, 1).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))
	mock.ExpectExec(`UPDATE tasks`).
		WithArgs(columnID, 2).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectCommit()

	if err := repo.DeleteTask(context.Background(), taskID, 1); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryDeleteTaskStaleVersionRollsBack(t *testing.T) {
	repo, mock := mockRepo(t)
	taskID, columnID := uuid.New(), uuid.New()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT column_id, position FROM tasks`).
		WithArgs(taskID).
		WillReturnRows(pgxmock.NewRows([]string{"column_id", "position"}).AddRow(columnID, 2))
	mock.ExpectExec(`DELETE FROM tasks`).
		WithArgs(taskID, 1).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))
	mock.ExpectQuery(guardExistence.String()).
		WithArgs(taskID).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectRollback()

	if err := repo.DeleteTask(context.Background(), taskID, 1); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
