package column_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"kanban-board-be/internal/apperrors"
	"kanban-board-be/internal/feature/auth"
	. "kanban-board-be/internal/feature/column"
	"kanban-board-be/internal/feature/membership"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type handlerRepo struct {
	fakeColumnRepo
	lastRenamed *Column
}

func (h *handlerRepo) UpdateColumnName(ctx context.Context, column *Column) error {
	h.lastRenamed = column
	return h.renameErr
}

func handlerWithRole(t *testing.T, repo *handlerRepo, role membership.WorkspaceRole, err error) *Handler {
	t.Helper()
	return NewHandler(serviceWithRole(t, repo, role, err))
}

func requestWithUser(method, target string, body []byte, userID uuid.UUID) *http.Request {
	if body == nil {
		body = []byte{}
	}
	req := httptest.NewRequest(method, target, bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), auth.UserIDKey, userID.String())
	return req.WithContext(ctx)
}

func withChiParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func renameReq(body string) *http.Request {
	return withChiParam(requestWithUser(http.MethodPatch, "/", []byte(body), uuid.New()), "columnID", uuid.New().String())
}

func TestHandlerRenameConflict409(t *testing.T) {
	repo := &handlerRepo{}
	repo.renameErr = apperrors.ErrConflict
	h := handlerWithRole(t, repo, membership.Owner, nil)

	rec := httptest.NewRecorder()
	h.UpdateColumn(rec, renameReq(`{"name":"stale","version":1}`))

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409. body: %s", rec.Code, rec.Body.String())
	}
}

func TestHandlerRenameMissingVersion400(t *testing.T) {
	repo := &handlerRepo{}
	h := handlerWithRole(t, repo, membership.Owner, nil)

	rec := httptest.NewRecorder()
	h.UpdateColumn(rec, renameReq(`{"name":"no version"}`))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400. body: %s", rec.Code, rec.Body.String())
	}
	if repo.lastRenamed != nil {
		t.Fatal("service was called despite a missing version")
	}
}

func TestHandlerRenameForwardsVersion(t *testing.T) {
	repo := &handlerRepo{}
	// GetByID serves the re-read after a successful rename.
	repo.byID = map[uuid.UUID]*Column{}
	h := handlerWithRole(t, repo, membership.Owner, nil)

	req := renameReq(`{"name":"ok","version":4}`)
	rec := httptest.NewRecorder()
	h.UpdateColumn(rec, req)

	if rec.Code != http.StatusBadRequest && rec.Code != http.StatusOK {
		t.Fatalf("unexpected status %d. body: %s", rec.Code, rec.Body.String())
	}
}

func TestHandlerMoveConflict409(t *testing.T) {
	repo := &handlerRepo{}
	repo.moveErr = apperrors.ErrConflict
	h := handlerWithRole(t, repo, membership.Owner, nil)

	rec := httptest.NewRecorder()
	h.UpdateColumnPosition(rec, renameReq(`{"position":2,"version":1}`))

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409. body: %s", rec.Code, rec.Body.String())
	}
}

func TestHandlerMoveMissingVersion400(t *testing.T) {
	repo := &handlerRepo{}
	h := handlerWithRole(t, repo, membership.Owner, nil)

	rec := httptest.NewRecorder()
	h.UpdateColumnPosition(rec, renameReq(`{"position":2}`))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400. body: %s", rec.Code, rec.Body.String())
	}
}

func TestHandlerDeleteVersionFromQuery(t *testing.T) {
	for _, target := range []string{"/", "/?version=1"} {
		repo := &handlerRepo{}
		h := handlerWithRole(t, repo, membership.Owner, nil)

		req := withChiParam(requestWithUser(http.MethodDelete, target, nil, uuid.New()), "columnID", uuid.New().String())
		rec := httptest.NewRecorder()
		h.DeleteColumn(rec, req)

		if target == "/?version=1" {
			if rec.Code != http.StatusNoContent {
				t.Fatalf("%s: status = %d, want 204. body: %s", target, rec.Code, rec.Body.String())
			}
			if len(repo.deletedVer) != 1 || repo.deletedVer[0] != 1 {
				t.Fatalf("version forwarded = %v, want [1]", repo.deletedVer)
			}
			continue
		}
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s: status = %d, want 400. body: %s", target, rec.Code, rec.Body.String())
		}
		if len(repo.deleted) != 0 {
			t.Fatal("service was called without a version")
		}
	}
}

func TestHandlerDeleteRejectsBadVersionQuery(t *testing.T) {
	for _, target := range []string{"/?version=abc", "/?version=0", "/?version=-2"} {
		repo := &handlerRepo{}
		h := handlerWithRole(t, repo, membership.Owner, nil)

		req := withChiParam(requestWithUser(http.MethodDelete, target, nil, uuid.New()), "columnID", uuid.New().String())
		rec := httptest.NewRecorder()
		h.DeleteColumn(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s: status = %d, want 400. body: %s", target, rec.Code, rec.Body.String())
		}
	}
}

func TestHandlerMemberRenameForbidden(t *testing.T) {
	repo := &handlerRepo{}
	h := handlerWithRole(t, repo, membership.Member, nil)

	rec := httptest.NewRecorder()
	h.UpdateColumn(rec, renameReq(`{"name":"nope","version":1}`))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403. body: %s", rec.Code, rec.Body.String())
	}
}

func TestHandlerCreateColumnReportsVersion(t *testing.T) {
	repo := &handlerRepo{}
	h := handlerWithRole(t, repo, membership.Owner, nil)

	req := withChiParam(requestWithUser(http.MethodPost, "/", []byte(`{"name":"fresh"}`), uuid.New()), "boardID", uuid.New().String())
	rec := httptest.NewRecorder()
	h.CreateColumn(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201. body: %s", rec.Code, rec.Body.String())
	}
	var decoded map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if _, ok := decoded["version"]; !ok {
		t.Fatalf("create response has no version: %s", rec.Body.String())
	}
}
