package task

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"kanban-board-be/internal/feature/membership"

	"github.com/google/uuid"
)

func TestHandlerCreateTaskWithMetadata(t *testing.T) {
	repo := &handlerRepo{}
	h := handlerWithRole(t, repo, membership.Member, nil)

	columnID := uuid.New()
	userID := uuid.New()
	body := []byte(`{"title":"meta","priority":"high","due_date":"2026-10-02","completed":true}`)

	req := withChiParam(requestWithUser(http.MethodPost, "/", body, userID), "columnID", columnID.String())
	rec := httptest.NewRecorder()
	h.CreateTask(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201. body: %s", rec.Code, rec.Body.String())
	}
	if repo.lastCreated == nil {
		t.Fatal("service was never called")
	}
	got := repo.lastCreated
	if got.ColumnID != columnID {
		t.Fatalf("column_id = %s, want %s", got.ColumnID, columnID)
	}
	if got.Priority == nil || *got.Priority != PriorityHigh {
		t.Fatalf("priority = %v, want high", got.Priority)
	}
	if d, ok := got.DueDate.Get(); !ok || d.Format(DateLayout) != "2026-10-02" {
		t.Fatalf("due_date = %v, want 2026-10-02", got.DueDate)
	}
	if got.Completed == nil || !*got.Completed {
		t.Fatalf("completed = %v, want true", got.Completed)
	}
}

func TestHandlerCreateTaskDefaultsWhenMetadataAbsent(t *testing.T) {
	repo := &handlerRepo{}
	h := handlerWithRole(t, repo, membership.Member, nil)

	columnID := uuid.New()
	body := []byte(`{"title":"plain"}`)

	req := withChiParam(requestWithUser(http.MethodPost, "/", body, uuid.New()), "columnID", columnID.String())
	rec := httptest.NewRecorder()
	h.CreateTask(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201. body: %s", rec.Code, rec.Body.String())
	}
	if repo.lastCreated == nil {
		t.Fatal("service was never called")
	}
	if repo.lastCreated.Priority != nil {
		t.Fatalf("priority = %v, want nil (DB default applies)", *repo.lastCreated.Priority)
	}
	if repo.lastCreated.DueDate.IsSet() {
		t.Fatal("due_date is set for an absent key, want unset")
	}
	if repo.lastCreated.Completed != nil {
		t.Fatalf("completed = %v, want nil (DB default applies)", *repo.lastCreated.Completed)
	}
}

func TestHandlerCreateTaskRejectsInvalidPriority(t *testing.T) {
	repo := &handlerRepo{}
	h := handlerWithRole(t, repo, membership.Member, nil)

	columnID := uuid.New()
	body := []byte(`{"title":"bad","priority":"urgent"}`)

	req := withChiParam(requestWithUser(http.MethodPost, "/", body, uuid.New()), "columnID", columnID.String())
	rec := httptest.NewRecorder()
	h.CreateTask(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400. body: %s", rec.Code, rec.Body.String())
	}
	if repo.lastCreated != nil {
		t.Fatal("repository was called despite invalid priority")
	}
}
