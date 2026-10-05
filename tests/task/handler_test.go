package task_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"kanban-board-be/internal/feature/auth"
	"kanban-board-be/internal/feature/membership"
	. "kanban-board-be/internal/feature/task"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type handlerRepo struct {
	fakeTaskRepo
	lastCreated *Task
	lastUpdated *Task
}

func (h *handlerRepo) Create(ctx context.Context, task *Task) error {
	h.lastCreated = task
	return nil
}

func (h *handlerRepo) UpdateTask(ctx context.Context, task *Task) error {
	h.lastUpdated = task
	return h.updateErr
}

func handlerWithRole(t *testing.T, repo *handlerRepo, role membership.WorkspaceRole, err error) *Handler {
	t.Helper()
	svc := serviceWithRole(t, repo, role, err)
	return NewHandler(svc, nil, nil)
}

func requestWithUser(method, target string, body []byte, userID uuid.UUID) *http.Request {
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader([]byte{})
	} else {
		reader = bytes.NewReader(body)
	}
	req := httptest.NewRequest(method, target, reader)
	ctx := context.WithValue(req.Context(), auth.UserIDKey, userID.String())
	return req.WithContext(ctx)
}

func withChiParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

// errNoRows mirrors pgx.ErrNoRows so tests hit the authz service's
// mapAccessError path (ErrNoRows -> ErrForbidden) through the real code.
func errNoRows() error { return pgx.ErrNoRows }
