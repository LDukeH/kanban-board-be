package column_test

import (
	"context"
	"testing"

	"kanban-board-be/internal/feature/authz"
	"kanban-board-be/internal/feature/board"
	. "kanban-board-be/internal/feature/column"
	"kanban-board-be/internal/feature/membership"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type fakeColumnRepo struct {
	renamed    []*Column
	moved      []*Column
	deleted    []uuid.UUID
	deletedVer []int
	createErr  error
	renameErr  error
	moveErr    error
	deleteErr  error
	byID       map[uuid.UUID]*Column
	byIDErr    error
}

func (f *fakeColumnRepo) Create(ctx context.Context, column *Column) error {
	return f.createErr
}

func (f *fakeColumnRepo) GetByBoardID(ctx context.Context, boardID uuid.UUID) ([]*Column, error) {
	return nil, nil
}

func (f *fakeColumnRepo) GetByID(ctx context.Context, columnID uuid.UUID) (*Column, error) {
	if f.byIDErr != nil {
		return nil, f.byIDErr
	}
	if c, ok := f.byID[columnID]; ok {
		return c, nil
	}
	return &Column{ID: columnID, Version: 1}, nil
}

func (f *fakeColumnRepo) UpdateColumnName(ctx context.Context, column *Column) error {
	if f.renameErr != nil {
		return f.renameErr
	}
	f.renamed = append(f.renamed, column)
	return nil
}

func (f *fakeColumnRepo) UpdateColumnPosition(ctx context.Context, columnID uuid.UUID, newPosition int, version int) error {
	if f.moveErr != nil {
		return f.moveErr
	}
	f.moved = append(f.moved, &Column{ID: columnID, Position: newPosition, Version: version})
	return nil
}

func (f *fakeColumnRepo) DeleteColumn(ctx context.Context, columnID uuid.UUID, version int) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	f.deleted = append(f.deleted, columnID)
	f.deletedVer = append(f.deletedVer, version)
	return nil
}

// columnRepo mirrors the service's storage contract so any test double can be passed in.
type columnRepo interface {
	Create(ctx context.Context, column *Column) error
	GetByBoardID(ctx context.Context, boardID uuid.UUID) ([]*Column, error)
	GetByID(ctx context.Context, columnID uuid.UUID) (*Column, error)
	UpdateColumnName(ctx context.Context, column *Column) error
	UpdateColumnPosition(ctx context.Context, columnID uuid.UUID, newPosition int, version int) error
	DeleteColumn(ctx context.Context, columnID uuid.UUID, version int) error
}

type fakeBoardRepo struct {
	board *board.Board
	err   error
}

func (f *fakeBoardRepo) GetByID(ctx context.Context, boardID uuid.UUID) (*board.Board, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.board != nil {
		return f.board, nil
	}
	return &board.Board{ID: boardID}, nil
}

// stubResolver implements authz.Resolver with a fixed role.
type stubResolver struct {
	role membership.WorkspaceRole
	err  error
}

func (s stubResolver) GetWorkspaceRole(ctx context.Context, userID, workspaceID uuid.UUID) (membership.WorkspaceRole, error) {
	return s.role, s.err
}

func (s stubResolver) GetBoardRole(ctx context.Context, userID, boardID uuid.UUID) (uuid.UUID, membership.WorkspaceRole, error) {
	return uuid.Nil, s.role, s.err
}

func (s stubResolver) GetColumnRole(ctx context.Context, userID, columnID uuid.UUID) (uuid.UUID, membership.WorkspaceRole, error) {
	return uuid.Nil, s.role, s.err
}

func (s stubResolver) GetTaskRole(ctx context.Context, userID, taskID uuid.UUID) (uuid.UUID, uuid.UUID, membership.WorkspaceRole, error) {
	return uuid.Nil, uuid.Nil, s.role, s.err
}

func (s stubResolver) GetLabelRole(ctx context.Context, userID, labelID uuid.UUID) (uuid.UUID, membership.WorkspaceRole, error) {
	return uuid.Nil, s.role, s.err
}

// errNoRows mirrors pgx.ErrNoRows so tests hit the authz ErrNoRows -> ErrForbidden path.
func errNoRows() error { return pgx.ErrNoRows }

func serviceWithRole(t *testing.T, repo columnRepo, role membership.WorkspaceRole, err error) *Service {
	t.Helper()
	authzSvc := authz.NewService(stubResolver{role: role, err: err})
	return NewService(&fakeBoardRepo{}, repo, authzSvc)
}
