package task_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"kanban-board-be/internal/apperrors"
	"kanban-board-be/internal/feature/membership"
	. "kanban-board-be/internal/feature/task"

	"github.com/google/uuid"
)

func patchTask(body string) *http.Request {
	req := requestWithUser(http.MethodPatch, "/", []byte(body), uuid.New())
	return withChiParam(req, "taskID", uuid.New().String())
}

func TestHandlerUpdateTaskConflict409(t *testing.T) {
	repo := &handlerRepo{}
	repo.updateErr = apperrors.ErrConflict
	h := handlerWithRole(t, repo, membership.Member, nil)

	rec := httptest.NewRecorder()
	h.UpdateTask(rec, patchTask(`{"title":"stale","version":1}`))

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409. body: %s", rec.Code, rec.Body.String())
	}
}

func TestHandlerUpdateTaskNotFound404(t *testing.T) {
	repo := &handlerRepo{}
	repo.updateErr = apperrors.ErrNotFound
	h := handlerWithRole(t, repo, membership.Member, nil)

	rec := httptest.NewRecorder()
	h.UpdateTask(rec, patchTask(`{"title":"gone","version":1}`))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404. body: %s", rec.Code, rec.Body.String())
	}
}

func TestHandlerUpdateTaskMissingVersion400(t *testing.T) {
	repo := &handlerRepo{}
	h := handlerWithRole(t, repo, membership.Member, nil)

	rec := httptest.NewRecorder()
	h.UpdateTask(rec, patchTask(`{"title":"no version"}`))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400. body: %s", rec.Code, rec.Body.String())
	}
	if repo.lastUpdated != nil {
		t.Fatal("service was called despite a missing version")
	}
}

func TestHandlerUpdateTaskForwardsVersion(t *testing.T) {
	repo := &handlerRepo{}
	h := handlerWithRole(t, repo, membership.Member, nil)

	rec := httptest.NewRecorder()
	h.UpdateTask(rec, patchTask(`{"title":"ok","version":7}`))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200. body: %s", rec.Code, rec.Body.String())
	}
	if repo.lastUpdated.Version != 7 {
		t.Fatalf("version forwarded = %d, want 7", repo.lastUpdated.Version)
	}
}

func TestHandlerUpdateTaskInvalidRoleForbidden(t *testing.T) {
	repo := &handlerRepo{}
	h := handlerWithRole(t, repo, membership.WorkspaceRole("viewer"), nil)

	rec := httptest.NewRecorder()
	h.UpdateTask(rec, patchTask(`{"title":"nope","version":1}`))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403. body: %s", rec.Code, rec.Body.String())
	}
}

func TestHandlerCreateTaskReportsVersion(t *testing.T) {
	repo := &handlerRepo{}
	h := handlerWithRole(t, repo, membership.Member, nil)

	body := []byte(`{"title":"fresh"}`)
	req := withChiParam(requestWithUser(http.MethodPost, "/", body, uuid.New()), "columnID", uuid.New().String())
	rec := httptest.NewRecorder()
	h.CreateTask(rec, req)

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

func moveReq(body string) *http.Request {
	req := requestWithUser(http.MethodPatch, "/", []byte(body), uuid.New())
	return withChiParam(req, "taskID", uuid.New().String())
}

func TestHandlerMoveConflict409(t *testing.T) {
	repo := &handlerRepo{}
	repo.positionErr = apperrors.ErrConflict
	h := handlerWithRole(t, repo, membership.Member, nil)

	rec := httptest.NewRecorder()
	h.UpdateTaskPosition(rec, moveReq(`{"column_id":"`+uuid.New().String()+`","position":2,"version":1}`))

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409. body: %s", rec.Code, rec.Body.String())
	}
}

func TestHandlerMoveMissingVersion400(t *testing.T) {
	repo := &handlerRepo{}
	h := handlerWithRole(t, repo, membership.Member, nil)

	rec := httptest.NewRecorder()
	h.UpdateTaskPosition(rec, moveReq(`{"column_id":"`+uuid.New().String()+`","position":2}`))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400. body: %s", rec.Code, rec.Body.String())
	}
}

func TestHandlerDeleteTaskVersionFromQuery(t *testing.T) {
	for _, target := range []string{"/", "/?version=1"} {
		repo := &handlerRepo{}
		h := handlerWithRole(t, repo, membership.Owner, nil)

		req := withChiParam(requestWithUser(http.MethodDelete, target, nil, uuid.New()), "taskID", uuid.New().String())
		rec := httptest.NewRecorder()
		h.DeleteTask(rec, req)

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

func TestHandlerDeleteTaskRejectsBadVersionQuery(t *testing.T) {
	for _, target := range []string{"/?version=abc", "/?version=0", "/?version=-3"} {
		repo := &handlerRepo{}
		h := handlerWithRole(t, repo, membership.Owner, nil)

		req := withChiParam(requestWithUser(http.MethodDelete, target, nil, uuid.New()), "taskID", uuid.New().String())
		rec := httptest.NewRecorder()
		h.DeleteTask(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s: status = %d, want 400. body: %s", target, rec.Code, rec.Body.String())
		}
	}
}

func TestHandlerDeleteTaskConflict409(t *testing.T) {
	repo := &handlerRepo{}
	repo.deleteErr = apperrors.ErrConflict
	h := handlerWithRole(t, repo, membership.Owner, nil)

	req := withChiParam(requestWithUser(http.MethodDelete, "/?version=1", nil, uuid.New()), "taskID", uuid.New().String())
	rec := httptest.NewRecorder()
	h.DeleteTask(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409. body: %s", rec.Code, rec.Body.String())
	}
}

func TestServiceDoesNotSwallowRepositoryErrors(t *testing.T) {
	cases := []struct {
		name string
		repo TaskRepository
		run  func(*Service) error
		want error
	}{
		{
			name: "update conflict",
			repo: &fakeTaskRepo{updateErr: apperrors.ErrConflict},
			run: func(s *Service) error {
				return s.UpdateTask(context.Background(), uuid.New(), &Task{ID: uuid.New(), Version: 1})
			},
			want: apperrors.ErrConflict,
		},
		{
			name: "update not found",
			repo: &fakeTaskRepo{updateErr: apperrors.ErrNotFound},
			run: func(s *Service) error {
				return s.UpdateTask(context.Background(), uuid.New(), &Task{ID: uuid.New(), Version: 1})
			},
			want: apperrors.ErrNotFound,
		},
		{
			name: "move conflict",
			repo: &fakeTaskRepo{positionErr: apperrors.ErrConflict},
			run: func(s *Service) error {
				return s.UpdateTaskPosition(context.Background(), uuid.New(), &Task{ID: uuid.New(), ColumnID: uuid.New(), Version: 1})
			},
			want: apperrors.ErrConflict,
		},
		{
			name: "delete conflict",
			repo: &fakeTaskRepo{deleteErr: apperrors.ErrConflict},
			run: func(s *Service) error {
				return s.DeleteTask(context.Background(), uuid.New(), uuid.New(), 1)
			},
			want: apperrors.ErrConflict,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := serviceWithRole(t, tc.repo, membership.Owner, nil)
			if err := tc.run(svc); !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestServiceRejectsMissingVersion(t *testing.T) {
	svc := serviceWithRole(t, &fakeTaskRepo{}, membership.Member, nil)

	err := svc.UpdateTask(context.Background(), uuid.New(), &Task{ID: uuid.New()})
	if !errors.Is(err, apperrors.ErrValidation) {
		t.Fatalf("update err = %v, want ErrValidation", err)
	}
	err = svc.UpdateTaskPosition(context.Background(), uuid.New(), &Task{ID: uuid.New(), ColumnID: uuid.New()})
	if !errors.Is(err, apperrors.ErrValidation) {
		t.Fatalf("move err = %v, want ErrValidation", err)
	}
	ownerSvc := serviceWithRole(t, &fakeTaskRepo{}, membership.Owner, nil)
	err = ownerSvc.DeleteTask(context.Background(), uuid.New(), uuid.New(), 0)
	if !errors.Is(err, apperrors.ErrValidation) {
		t.Fatalf("delete err = %v, want ErrValidation", err)
	}
}

func TestServiceMemberStillMovesAndUpdates(t *testing.T) {
	repo := &fakeTaskRepo{}
	svc := serviceWithRole(t, repo, membership.Member, nil)

	if err := svc.UpdateTask(context.Background(), uuid.New(), &Task{ID: uuid.New(), Title: strp("t"), Version: 3}); err != nil {
		t.Fatalf("member update: %v", err)
	}
	if len(repo.updated) != 1 || repo.updated[0].Version != 3 {
		t.Fatalf("updated = %v, want one task with version 3", repo.updated)
	}

	if err := svc.UpdateTaskPosition(context.Background(), uuid.New(), &Task{ID: uuid.New(), ColumnID: uuid.New(), Position: 1, Version: 2}); err != nil {
		t.Fatalf("member move: %v", err)
	}
	if len(repo.moved) != 1 || repo.moved[0].Version != 2 {
		t.Fatalf("moved = %v, want one task with version 2", repo.moved)
	}
}
