package task

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"kanban-board-be/internal/feature/membership"

	"github.com/google/uuid"
)

func TestHandlerUpdateTaskDueDateNullClears(t *testing.T) {
	repo := &handlerRepo{}
	h := handlerWithRole(t, repo, membership.Member, nil)

	taskID := uuid.New()
	body := []byte(`{"due_date":null}`)

	req := withChiParam(requestWithUser(http.MethodPatch, "/", body, uuid.New()), "taskID", taskID.String())
	rec := httptest.NewRecorder()
	h.UpdateTask(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204. body: %s", rec.Code, rec.Body.String())
	}
	if repo.lastUpdated == nil {
		t.Fatal("service was never called")
	}
	if repo.lastUpdated.DueDate.State() != OptionalNull {
		t.Fatalf("due_date state = %v, want OptionalNull", repo.lastUpdated.DueDate.State())
	}
	// Only due_date was sent: no other field may render into the SET clause.
	if clauses, _ := buildTaskUpdateSet(repo.lastUpdated); len(clauses) != 1 || clauses[0] != "due_date = NULL" {
		t.Fatalf("clauses = %v, want [due_date = NULL]", clauses)
	}
}

func TestHandlerUpdateTaskDueDateAbsentLeavesUnchanged(t *testing.T) {
	repo := &handlerRepo{}
	h := handlerWithRole(t, repo, membership.Member, nil)

	taskID := uuid.New()
	body := []byte(`{"title":"renamed"}`)

	req := withChiParam(requestWithUser(http.MethodPatch, "/", body, uuid.New()), "taskID", taskID.String())
	rec := httptest.NewRecorder()
	h.UpdateTask(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204. body: %s", rec.Code, rec.Body.String())
	}
	if repo.lastUpdated == nil {
		t.Fatal("service was never called")
	}
	if repo.lastUpdated.DueDate.IsSet() {
		t.Fatal("due_date is set for an absent key, want unset")
	}
}

func TestHandlerUpdateTaskDueDateValueSets(t *testing.T) {
	repo := &handlerRepo{}
	h := handlerWithRole(t, repo, membership.Member, nil)

	taskID := uuid.New()
	body := []byte(`{"due_date":"2026-12-31"}`)

	req := withChiParam(requestWithUser(http.MethodPatch, "/", body, uuid.New()), "taskID", taskID.String())
	rec := httptest.NewRecorder()
	h.UpdateTask(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204. body: %s", rec.Code, rec.Body.String())
	}
	if repo.lastUpdated == nil {
		t.Fatal("service was never called")
	}
	if d, ok := repo.lastUpdated.DueDate.Get(); !ok || d.Format(DateLayout) != "2026-12-31" {
		t.Fatalf("due_date = %v, want 2026-12-31", repo.lastUpdated.DueDate)
	}
}

func TestHandlerUpdateTaskRejectsInvalidPriority(t *testing.T) {
	repo := &handlerRepo{}
	h := handlerWithRole(t, repo, membership.Member, nil)

	taskID := uuid.New()
	body := []byte(`{"priority":"whenever"}`)

	req := withChiParam(requestWithUser(http.MethodPatch, "/", body, uuid.New()), "taskID", taskID.String())
	rec := httptest.NewRecorder()
	h.UpdateTask(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400. body: %s", rec.Code, rec.Body.String())
	}
}

func TestHandlerUpdateTaskRejectsBadDueDate(t *testing.T) {
	repo := &handlerRepo{}
	h := handlerWithRole(t, repo, membership.Member, nil)

	taskID := uuid.New()
	body := []byte(`{"due_date":"tomorrow"}`)

	req := withChiParam(requestWithUser(http.MethodPatch, "/", body, uuid.New()), "taskID", taskID.String())
	rec := httptest.NewRecorder()
	h.UpdateTask(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400. body: %s", rec.Code, rec.Body.String())
	}
	if repo.lastUpdated != nil {
		t.Fatal("service was called despite a bad due_date")
	}
}

func TestHandlerMemberForbiddenSurfaces403(t *testing.T) {
	// The resolver itself denies access; authz maps pgx.ErrNoRows to
	// ErrForbidden, which the handler writes as 403.
	repo := &handlerRepo{}
	authzSvc := newAuthzServiceForTest(t, stubResolver{err: errNoRows()})
	svc := NewService(repo, authzSvc, nil, nil)
	h := NewHandler(svc, nil, nil)

	taskID := uuid.New()
	body := []byte(`{"title":"nope"}`)

	req := withChiParam(requestWithUser(http.MethodPatch, "/", body, uuid.New()), "taskID", taskID.String())
	rec := httptest.NewRecorder()
	h.UpdateTask(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403. body: %s", rec.Code, rec.Body.String())
	}
}

func TestHandlerResponseCarriesMetadataFields(t *testing.T) {
	repo := &handlerRepo{}
	h := handlerWithRole(t, repo, membership.Member, nil)

	columnID := uuid.New()
	body := []byte(`{"title":"shape"}`)

	req := withChiParam(requestWithUser(http.MethodPost, "/", body, uuid.New()), "columnID", columnID.String())
	rec := httptest.NewRecorder()
	h.CreateTask(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}
	var decoded map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	for _, key := range []string{"priority", "due_date", "completed"} {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("response missing key %q: %s", key, rec.Body.String())
		}
	}
}
