package column_test

import (
	"context"
	"errors"
	"testing"

	"kanban-board-be/internal/apperrors"
	"kanban-board-be/internal/feature/auth"
	. "kanban-board-be/internal/feature/column"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
)

func mockRepo(t *testing.T) (*Repository, pgxmock.PgxPoolIface) {
	t.Helper()
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool: %v", err)
	}
	return NewRepositoryWithDB(mock), mock
}

// actorCtx returns a context carrying the acting user's ID, the way the auth
// middleware does for production requests.
func actorCtx() (context.Context, uuid.UUID) {
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), auth.UserIDKey, userID.String())
	return ctx, userID
}

// expectActor queues the set_config call that stamps app.actor_id at the start
// of every mutation transaction.
func expectActor(mock pgxmock.PgxPoolIface, userID uuid.UUID) {
	mock.ExpectExec(`set_config`).
		WithArgs(userID.String()).
		WillReturnResult(pgxmock.NewResult("SELECT", 1))
}

func TestRepositoryRenameSuccess(t *testing.T) {
	repo, mock := mockRepo(t)
	columnID := uuid.New()
	ctx, userID := actorCtx()

	mock.ExpectBegin()
	expectActor(mock, userID)
	mock.ExpectExec(`UPDATE columns`).
		WithArgs("renamed", columnID, 1).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectCommit()

	column := &Column{ID: columnID, Name: "renamed", Version: 1}
	if err := repo.UpdateColumnName(ctx, column); err != nil {
		t.Fatalf("UpdateColumnName: %v", err)
	}
	if column.Version != 2 {
		t.Fatalf("version = %d, want 2", column.Version)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// Two clients read version 1. The first write wins; the second is refused.
func TestRepositoryRenameTwoClientsSameVersion(t *testing.T) {
	repo, mock := mockRepo(t)
	columnID := uuid.New()
	ctx, userID := actorCtx()

	mock.ExpectBegin()
	expectActor(mock, userID)
	mock.ExpectExec(`UPDATE columns`).
		WithArgs("first", columnID, 1).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	expectActor(mock, userID)
	mock.ExpectExec(`UPDATE columns`).
		WithArgs("second", columnID, 1).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))
	mock.ExpectQuery(`SELECT EXISTS \(SELECT 1 FROM columns WHERE id = \$1\)`).
		WithArgs(columnID).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectRollback()

	first := &Column{ID: columnID, Name: "first", Version: 1}
	if err := repo.UpdateColumnName(ctx, first); err != nil {
		t.Fatalf("first: %v", err)
	}
	if first.Version != 2 {
		t.Fatalf("first version = %d, want 2", first.Version)
	}

	second := &Column{ID: columnID, Name: "second", Version: 1}
	if err := repo.UpdateColumnName(ctx, second); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("second = %v, want ErrConflict", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryRenameMissingRowIsNotFound(t *testing.T) {
	repo, mock := mockRepo(t)
	columnID := uuid.New()
	ctx, userID := actorCtx()

	mock.ExpectBegin()
	expectActor(mock, userID)
	mock.ExpectExec(`UPDATE columns`).
		WithArgs("t", columnID, 1).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))
	mock.ExpectQuery(`SELECT EXISTS \(SELECT 1 FROM columns WHERE id = \$1\)`).
		WithArgs(columnID).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectRollback()

	err := repo.UpdateColumnName(ctx, &Column{ID: columnID, Name: "t", Version: 1})
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryMoveSuccess(t *testing.T) {
	repo, mock := mockRepo(t)
	columnID, boardID := uuid.New(), uuid.New()
	ctx, userID := actorCtx()

	mock.ExpectBegin()
	expectActor(mock, userID)
	mock.ExpectQuery(`SELECT board_id, position FROM columns WHERE id = \$1 FOR UPDATE`).
		WithArgs(columnID).
		WillReturnRows(pgxmock.NewRows([]string{"board_id", "position"}).AddRow(boardID, 0))
	mock.ExpectExec(`UPDATE columns`).
		WithArgs(boardID, 0, 2).
		WillReturnResult(pgxmock.NewResult("UPDATE", 2))
	mock.ExpectExec(`UPDATE columns`).
		WithArgs(2, columnID, 1).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectCommit()

	if err := repo.UpdateColumnPosition(ctx, columnID, 2, 1); err != nil {
		t.Fatalf("UpdateColumnPosition: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// A stale destination must roll the whole reorder back. ExpectCommit is
// deliberately not queued, so any commit call fails the test.
func TestRepositoryMoveStaleVersionRollsBack(t *testing.T) {
	repo, mock := mockRepo(t)
	columnID, boardID := uuid.New(), uuid.New()
	ctx, userID := actorCtx()

	mock.ExpectBegin()
	expectActor(mock, userID)
	mock.ExpectQuery(`SELECT board_id, position FROM columns WHERE id = \$1 FOR UPDATE`).
		WithArgs(columnID).
		WillReturnRows(pgxmock.NewRows([]string{"board_id", "position"}).AddRow(boardID, 0))
	mock.ExpectExec(`UPDATE columns`).
		WithArgs(boardID, 0, 2).
		WillReturnResult(pgxmock.NewResult("UPDATE", 2))
	mock.ExpectExec(`UPDATE columns`).
		WithArgs(2, columnID, 1).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))
	mock.ExpectQuery(`SELECT EXISTS \(SELECT 1 FROM columns WHERE id = \$1\)`).
		WithArgs(columnID).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectRollback()

	if err := repo.UpdateColumnPosition(ctx, columnID, 2, 1); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryDeleteSuccess(t *testing.T) {
	repo, mock := mockRepo(t)
	columnID, boardID := uuid.New(), uuid.New()
	ctx, userID := actorCtx()

	mock.ExpectBegin()
	expectActor(mock, userID)
	mock.ExpectQuery(`SELECT board_id, position FROM columns`).
		WithArgs(columnID).
		WillReturnRows(pgxmock.NewRows([]string{"board_id", "position"}).AddRow(boardID, 1))
	mock.ExpectExec(`DELETE FROM columns`).
		WithArgs(columnID, 1).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))
	mock.ExpectExec(`UPDATE columns`).
		WithArgs(boardID, 1).
		WillReturnResult(pgxmock.NewResult("UPDATE", 2))
	mock.ExpectCommit()

	if err := repo.DeleteColumn(ctx, columnID, 1); err != nil {
		t.Fatalf("DeleteColumn: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryDeleteStaleVersionRollsBack(t *testing.T) {
	repo, mock := mockRepo(t)
	columnID, boardID := uuid.New(), uuid.New()
	ctx, userID := actorCtx()

	mock.ExpectBegin()
	expectActor(mock, userID)
	mock.ExpectQuery(`SELECT board_id, position FROM columns`).
		WithArgs(columnID).
		WillReturnRows(pgxmock.NewRows([]string{"board_id", "position"}).AddRow(boardID, 1))
	mock.ExpectExec(`DELETE FROM columns`).
		WithArgs(columnID, 1).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))
	mock.ExpectQuery(`SELECT EXISTS \(SELECT 1 FROM columns WHERE id = \$1\)`).
		WithArgs(columnID).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectRollback()

	if err := repo.DeleteColumn(ctx, columnID, 1); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
