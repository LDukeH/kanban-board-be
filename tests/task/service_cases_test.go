package task_test

import (
	"context"
	"errors"
	"testing"

	"kanban-board-be/internal/apperrors"
	"kanban-board-be/internal/feature/membership"
	. "kanban-board-be/internal/feature/task"

	"github.com/google/uuid"
)

func TestServiceCreateTaskCarriesMetadata(t *testing.T) {
	repo := &fakeTaskRepo{}
	svc := serviceWithRole(t, repo, membership.Member, nil)

	completed := true
	task := &Task{
		ColumnID:  uuid.New(),
		Title:     strp("with metadata"),
		Priority:  strp(PriorityMedium),
		DueDate:   Some(mustParseDate(t, "2026-10-02")),
		Completed: &completed,
	}
	if err := svc.CreateTask(context.Background(), uuid.New(), task); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if len(repo.created) != 1 {
		t.Fatalf("created = %d tasks, want 1", len(repo.created))
	}
	got := repo.created[0]
	if got.Priority == nil || *got.Priority != PriorityMedium {
		t.Fatalf("priority = %v, want medium", got.Priority)
	}
	if d, ok := got.DueDate.Get(); !ok || d.Format(DateLayout) != "2026-10-02" {
		t.Fatalf("due_date = %v, want 2026-10-02", got.DueDate)
	}
	if got.Completed == nil || !*got.Completed {
		t.Fatalf("completed = %v, want true", got.Completed)
	}
}

func TestServiceCreateTaskRejectsInvalidPriority(t *testing.T) {
	repo := &fakeTaskRepo{}
	svc := serviceWithRole(t, repo, membership.Member, nil)

	task := &Task{ColumnID: uuid.New(), Title: strp("bad"), Priority: strp("urgent")}
	err := svc.CreateTask(context.Background(), uuid.New(), task)
	if !errors.Is(err, apperrors.ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
	if len(repo.created) != 0 {
		t.Fatal("repository was called despite invalid priority")
	}
}

func TestServiceUpdateTaskRejectsInvalidPriority(t *testing.T) {
	repo := &fakeTaskRepo{}
	svc := serviceWithRole(t, repo, membership.Member, nil)

	task := &Task{ID: uuid.New(), Priority: strp("HIGH")}
	err := svc.UpdateTask(context.Background(), uuid.New(), task)
	if !errors.Is(err, apperrors.ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
	if len(repo.updated) != 0 {
		t.Fatal("repository was called despite invalid priority")
	}
}

func TestServiceMemberCanUpdateMetadata(t *testing.T) {
	repo := &fakeTaskRepo{}
	svc := serviceWithRole(t, repo, membership.Member, nil)

	task := &Task{ID: uuid.New(), Priority: strp(PriorityLow), Completed: boolp(true)}
	if err := svc.UpdateTask(context.Background(), uuid.New(), task); err != nil {
		t.Fatalf("Member UpdateTask: %v", err)
	}
	if len(repo.updated) != 1 {
		t.Fatalf("updated = %d tasks, want 1", len(repo.updated))
	}
}

func TestServiceMemberCannotDeleteTask(t *testing.T) {
	repo := &fakeTaskRepo{}
	svc := serviceWithRole(t, repo, membership.Member, nil)

	err := svc.DeleteTask(context.Background(), uuid.New(), uuid.New(), 1)
	if !errors.Is(err, apperrors.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
	if len(repo.deleted) != 0 {
		t.Fatal("repository deleted despite Member role")
	}
}

func TestServiceOwnerCanDeleteTask(t *testing.T) {
	repo := &fakeTaskRepo{}
	svc := serviceWithRole(t, repo, membership.Owner, nil)

	taskID := uuid.New()
	if err := svc.DeleteTask(context.Background(), uuid.New(), taskID, 1); err != nil {
		t.Fatalf("Owner DeleteTask: %v", err)
	}
	if len(repo.deleted) != 1 || repo.deleted[0] != taskID {
		t.Fatalf("deleted = %v, want [%s]", repo.deleted, taskID)
	}
}

func TestServiceOutsiderDenied(t *testing.T) {
	repo := &fakeTaskRepo{}
	svc := serviceWithRole(t, repo, membership.WorkspaceRole("outsider"), nil)

	if err := svc.CreateTask(context.Background(), uuid.New(), &Task{ColumnID: uuid.New()}); !errors.Is(err, apperrors.ErrForbidden) {
		t.Fatalf("create err = %v, want ErrForbidden", err)
	}
	if err := svc.UpdateTask(context.Background(), uuid.New(), &Task{ID: uuid.New()}); !errors.Is(err, apperrors.ErrForbidden) {
		t.Fatalf("update err = %v, want ErrForbidden", err)
	}
}
