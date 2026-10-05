package column_test

import (
	"context"
	"errors"
	"testing"

	"kanban-board-be/internal/apperrors"
	. "kanban-board-be/internal/feature/column"
	"kanban-board-be/internal/feature/membership"

	"github.com/google/uuid"
)

func TestClassifyVersionMismatch(t *testing.T) {
	if err := ClassifyVersionMismatch(true); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("exists -> %v, want ErrConflict", err)
	}
	if err := ClassifyVersionMismatch(false); !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("missing -> %v, want ErrNotFound", err)
	}
}

func TestServiceDoesNotSwallowRepositoryErrors(t *testing.T) {
	cases := []struct {
		name string
		repo *fakeColumnRepo
		run  func(*Service) error
		want error
	}{
		{
			name: "rename conflict",
			repo: &fakeColumnRepo{renameErr: apperrors.ErrConflict},
			run: func(s *Service) error {
				return s.UpdateColumnName(context.Background(), uuid.New(), &Column{ID: uuid.New(), Name: "n", Version: 1})
			},
			want: apperrors.ErrConflict,
		},
		{
			name: "rename not found",
			repo: &fakeColumnRepo{renameErr: apperrors.ErrNotFound},
			run: func(s *Service) error {
				return s.UpdateColumnName(context.Background(), uuid.New(), &Column{ID: uuid.New(), Name: "n", Version: 1})
			},
			want: apperrors.ErrNotFound,
		},
		{
			name: "move conflict",
			repo: &fakeColumnRepo{moveErr: apperrors.ErrConflict},
			run: func(s *Service) error {
				return s.UpdateColumnPosition(context.Background(), uuid.New(), &Column{ID: uuid.New(), Position: 1, Version: 1})
			},
			want: apperrors.ErrConflict,
		},
		{
			name: "delete conflict",
			repo: &fakeColumnRepo{deleteErr: apperrors.ErrConflict},
			run: func(s *Service) error {
				return s.DeleteColumn(context.Background(), uuid.New(), uuid.New(), 1)
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
	svc := serviceWithRole(t, &fakeColumnRepo{}, membership.Admin, nil)

	err := svc.UpdateColumnName(context.Background(), uuid.New(), &Column{ID: uuid.New(), Name: "n"})
	if !errors.Is(err, apperrors.ErrValidation) {
		t.Fatalf("rename err = %v, want ErrValidation", err)
	}
	err = svc.UpdateColumnPosition(context.Background(), uuid.New(), &Column{ID: uuid.New(), Position: 1})
	if !errors.Is(err, apperrors.ErrValidation) {
		t.Fatalf("move err = %v, want ErrValidation", err)
	}
	err = svc.DeleteColumn(context.Background(), uuid.New(), uuid.New(), 0)
	if !errors.Is(err, apperrors.ErrValidation) {
		t.Fatalf("delete err = %v, want ErrValidation", err)
	}
}

func TestServiceForwardsVersion(t *testing.T) {
	repo := &fakeColumnRepo{}
	svc := serviceWithRole(t, repo, membership.Admin, nil)

	if err := svc.UpdateColumnName(context.Background(), uuid.New(), &Column{ID: uuid.New(), Name: "n", Version: 5}); err != nil {
		t.Fatalf("rename: %v", err)
	}
	if len(repo.renamed) != 1 || repo.renamed[0].Version != 5 {
		t.Fatalf("renamed = %v, want one column with version 5", repo.renamed)
	}

	if err := svc.UpdateColumnPosition(context.Background(), uuid.New(), &Column{ID: uuid.New(), Position: 2, Version: 6}); err != nil {
		t.Fatalf("move: %v", err)
	}
	if len(repo.moved) != 1 || repo.moved[0].Version != 6 {
		t.Fatalf("moved = %v, want one column with version 6", repo.moved)
	}

	if err := svc.DeleteColumn(context.Background(), uuid.New(), uuid.New(), 7); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if len(repo.deletedVer) != 1 || repo.deletedVer[0] != 7 {
		t.Fatalf("deletedVer = %v, want [7]", repo.deletedVer)
	}
}

// Member may work on tasks but not on columns: the column rules are unchanged.
func TestServiceMemberCannotMutateColumn(t *testing.T) {
	repo := &fakeColumnRepo{}
	svc := serviceWithRole(t, repo, membership.Member, nil)

	if err := svc.UpdateColumnName(context.Background(), uuid.New(), &Column{ID: uuid.New(), Name: "n", Version: 1}); !errors.Is(err, apperrors.ErrForbidden) {
		t.Fatalf("rename err = %v, want ErrForbidden", err)
	}
	if err := svc.UpdateColumnPosition(context.Background(), uuid.New(), &Column{ID: uuid.New(), Position: 1, Version: 1}); !errors.Is(err, apperrors.ErrForbidden) {
		t.Fatalf("move err = %v, want ErrForbidden", err)
	}
	if err := svc.DeleteColumn(context.Background(), uuid.New(), uuid.New(), 1); !errors.Is(err, apperrors.ErrForbidden) {
		t.Fatalf("delete err = %v, want ErrForbidden", err)
	}
	if len(repo.renamed)+len(repo.moved)+len(repo.deleted) != 0 {
		t.Fatal("repository was called despite a forbidden role")
	}
}

func TestServiceOutsiderDenied(t *testing.T) {
	repo := &fakeColumnRepo{}
	svc := serviceWithRole(t, repo, membership.WorkspaceRole("viewer"), nil)

	if err := svc.UpdateColumnName(context.Background(), uuid.New(), &Column{ID: uuid.New(), Name: "n", Version: 1}); !errors.Is(err, apperrors.ErrForbidden) {
		t.Fatalf("rename err = %v, want ErrForbidden", err)
	}
}

func TestServiceInaccessibleColumnForbidden(t *testing.T) {
	svc := serviceWithRole(t, &fakeColumnRepo{}, "", errNoRows())

	if err := svc.DeleteColumn(context.Background(), uuid.New(), uuid.New(), 1); !errors.Is(err, apperrors.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}
