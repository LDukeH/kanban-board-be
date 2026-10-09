package task_test

import (
	"errors"
	"testing"

	"kanban-board-be/internal/apperrors"
	. "kanban-board-be/internal/feature/task"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
)

// expectCurrentTask queues the lookup that opens a move or delete.
func expectCurrentTask(mock pgxmock.PgxPoolIface, taskID, columnID uuid.UUID, position int) {
	mock.ExpectQuery(`SELECT id, column_id, position, version FROM tasks`).
		WithArgs(taskID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "column_id", "position", "version"}).AddRow(taskID, columnID, position, 1))
}

func TestRepositoryMoveNoOpUsesCurrentVersion(t *testing.T) {
	repo, mock := mockRepo(t)
	taskID, columnID := uuid.New(), uuid.New()
	ctx, userID := actorCtx()

	mock.ExpectBegin()
	expectActor(mock, userID)
	mock.ExpectQuery(`SELECT id, column_id, position, version FROM tasks`).
		WithArgs(taskID).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "column_id", "position", "version"}).
				AddRow(taskID, columnID, 2, 2),
		)
	mock.ExpectRollback()

	task := &Task{ID: taskID, ColumnID: columnID, Position: 2, Version: 1}
	err := repo.UpdateTaskPosition(ctx, task)
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if task.Version != 2 {
		t.Fatalf("version = %d, want current version 2", task.Version)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryMoveNoOpMatchingVersionSucceeds(t *testing.T) {
	repo, mock := mockRepo(t)
	taskID, columnID := uuid.New(), uuid.New()
	ctx, userID := actorCtx()

	mock.ExpectBegin()
	expectActor(mock, userID)
	mock.ExpectQuery(`SELECT id, column_id, position, version FROM tasks`).
		WithArgs(taskID).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "column_id", "position", "version"}).
				AddRow(taskID, columnID, 2, 1),
		)
	mock.ExpectRollback()

	err := repo.UpdateTaskPosition(
		ctx,
		&Task{ID: taskID, ColumnID: columnID, Position: 2, Version: 1},
	)
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryMoveSameColumnSuccess(t *testing.T) {
	repo, mock := mockRepo(t)
	taskID, columnID := uuid.New(), uuid.New()
	ctx, userID := actorCtx()

	mock.ExpectBegin()
	expectActor(mock, userID)
	expectCurrentTask(mock, taskID, columnID, 1)
	// Downwards: the tasks between old and new move up one slot.
	mock.ExpectExec(`UPDATE tasks`).
		WithArgs(columnID, 1, 3).
		WillReturnResult(pgxmock.NewResult("UPDATE", 2))
	mock.ExpectExec(`UPDATE tasks`).
		WithArgs(3, columnID, taskID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectCommit()

	task := &Task{ID: taskID, ColumnID: columnID, Position: 3, Version: 1}
	if err := repo.UpdateTaskPosition(ctx, task); err != nil {
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
	ctx, userID := actorCtx()

	mock.ExpectBegin()
	expectActor(mock, userID)
	expectCurrentTask(mock, taskID, sourceID, 0)
	mock.ExpectExec(`UPDATE tasks`).
		WithArgs(sourceID, 0).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectExec(`UPDATE tasks`).
		WithArgs(targetID, 2).
		WillReturnResult(pgxmock.NewResult("UPDATE", 3))
	mock.ExpectExec(`UPDATE tasks`).
		WithArgs(2, targetID, taskID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectCommit()

	task := &Task{ID: taskID, ColumnID: targetID, Position: 2, Version: 1}
	if err := repo.UpdateTaskPosition(ctx, task); err != nil {
		t.Fatalf("UpdateTaskPosition: %v", err)
	}
	if task.Version != 2 {
		t.Fatalf("version = %d, want 2", task.Version)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryMoveStaleVersionStillSucceeds(t *testing.T) {
	repo, mock := mockRepo(t)
	taskID, columnID := uuid.New(), uuid.New()
	ctx, userID := actorCtx()

	mock.ExpectBegin()
	expectActor(mock, userID)
	expectCurrentTask(mock, taskID, columnID, 1)
	mock.ExpectExec(`UPDATE tasks`).
		WithArgs(columnID, 1, 3).
		WillReturnResult(pgxmock.NewResult("UPDATE", 2))
	mock.ExpectExec(`UPDATE tasks`).
		WithArgs(3, columnID, taskID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectCommit()

	task := &Task{ID: taskID, ColumnID: columnID, Position: 3, Version: 1}
	if err := repo.UpdateTaskPosition(ctx, task); err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if task.Version != 2 {
		t.Fatalf("version = %d, want 2", task.Version)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryMoveStaleVersionCrossColumnStillSucceeds(t *testing.T) {
	repo, mock := mockRepo(t)
	taskID, sourceID, targetID := uuid.New(), uuid.New(), uuid.New()
	ctx, userID := actorCtx()

	mock.ExpectBegin()
	expectActor(mock, userID)
	expectCurrentTask(mock, taskID, sourceID, 0)
	mock.ExpectExec(`UPDATE tasks`).
		WithArgs(sourceID, 0).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectExec(`UPDATE tasks`).
		WithArgs(targetID, 2).
		WillReturnResult(pgxmock.NewResult("UPDATE", 3))
	mock.ExpectExec(`UPDATE tasks`).
		WithArgs(2, targetID, taskID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectCommit()

	task := &Task{ID: taskID, ColumnID: targetID, Position: 2, Version: 1}
	if err := repo.UpdateTaskPosition(ctx, task); err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if task.Version != 2 {
		t.Fatalf("version = %d, want 2", task.Version)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryDeleteTaskSuccess(t *testing.T) {
	repo, mock := mockRepo(t)
	taskID, columnID := uuid.New(), uuid.New()
	ctx, userID := actorCtx()

	mock.ExpectBegin()
	expectActor(mock, userID)
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

	if err := repo.DeleteTask(ctx, taskID, 1); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryDeleteTaskStaleVersionRollsBack(t *testing.T) {
	repo, mock := mockRepo(t)
	taskID, columnID := uuid.New(), uuid.New()
	ctx, userID := actorCtx()

	mock.ExpectBegin()
	expectActor(mock, userID)
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

	if err := repo.DeleteTask(ctx, taskID, 1); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
