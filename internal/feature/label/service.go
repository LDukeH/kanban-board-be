package label

import (
	"context"
	"kanban-board-be/internal/apperrors"
	"kanban-board-be/internal/feature/authz"
	"kanban-board-be/internal/feature/membership"

	"github.com/google/uuid"
)

type labelRepository interface {
	Create(ctx context.Context, label *Label) error
	GetLabelsByBoardID(ctx context.Context, boardID uuid.UUID) ([]*Label, error)
	GetLabelByID(ctx context.Context, labelID uuid.UUID) (*Label, error)
	UpdateLabel(ctx context.Context, label *Label) error
	DeleteLabel(ctx context.Context, labelID uuid.UUID) error
	AddLabelToTask(ctx context.Context, taskID uuid.UUID, labelID uuid.UUID) error
	RemoveLabelFromTask(ctx context.Context, taskID uuid.UUID, labelID uuid.UUID) error
	GetLabelsByTaskID(ctx context.Context, taskID uuid.UUID) ([]*Label, error)
}

type Service struct {
	repo  labelRepository
	authz *authz.Service
}

func NewService(repo labelRepository, authz *authz.Service) *Service {
	return &Service{
		repo:  repo,
		authz: authz,
	}
}

func (s *Service) CreateLabel(ctx context.Context, userID uuid.UUID, label *Label) error {
	role, err := s.authz.GetBoardRole(ctx, userID, label.BoardID)
	if err != nil {
		return err
	}
	if role != membership.Owner && role != membership.Admin {
		return apperrors.ErrForbidden
	}

	return s.repo.Create(ctx, label)
}

func (s *Service) GetLabelsByBoardID(ctx context.Context, userID uuid.UUID, boardID uuid.UUID) ([]*Label, error) {
	role, err := s.authz.GetBoardRole(ctx, userID, boardID)
	if err != nil {
		return nil, err
	}
	if role != membership.Owner && role != membership.Admin && role != membership.Member {
		return nil, apperrors.ErrForbidden
	}
	return s.repo.GetLabelsByBoardID(ctx, boardID)
}

func (s *Service) GetLabelByID(ctx context.Context, userID uuid.UUID, labelID uuid.UUID) (*Label, error) {
	role, err := s.authz.GetLabelRole(ctx, userID, labelID)
	if err != nil {
		return nil, err
	}
	if role != membership.Owner && role != membership.Admin && role != membership.Member {
		return nil, apperrors.ErrForbidden
	}

	return s.repo.GetLabelByID(ctx, labelID)
}

func (s *Service) UpdateLabel(ctx context.Context, userID uuid.UUID, label *Label) error {
	role, err := s.authz.GetLabelRole(ctx, userID, label.ID)
	if err != nil {
		return err
	}
	if role != membership.Owner && role != membership.Admin {
		return apperrors.ErrForbidden
	}
	return s.repo.UpdateLabel(ctx, label)
}

func (s *Service) DeleteLabel(ctx context.Context, userID uuid.UUID, labelID uuid.UUID) error {
	role, err := s.authz.GetLabelRole(ctx, userID, labelID)
	if err != nil {
		return err
	}
	if role != membership.Owner && role != membership.Admin {
		return apperrors.ErrForbidden
	}
	return s.repo.DeleteLabel(ctx, labelID)
}

func (s *Service) AddLabelToTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID, labelID uuid.UUID) error {
	role, err := s.authz.GetLabelRole(ctx, userID, labelID)
	if err != nil {
		return err
	}
	if role != membership.Owner && role != membership.Admin && role != membership.Member {
		return apperrors.ErrForbidden
	}

	return s.repo.AddLabelToTask(ctx, taskID, labelID)
}

func (s *Service) RemoveLabelFromTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID, labelID uuid.UUID) error {
	role, err := s.authz.GetLabelRole(ctx, userID, labelID)
	if err != nil {
		return err
	}
	if role != membership.Owner && role != membership.Admin && role != membership.Member {
		return apperrors.ErrForbidden
	}

	return s.repo.RemoveLabelFromTask(ctx, taskID, labelID)
}

func (s *Service) GetLabelsByTaskID(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) ([]*Label, error) {
	role, err := s.authz.GetTaskRole(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}
	if role != membership.Owner && role != membership.Admin && role != membership.Member {
		return nil, apperrors.ErrForbidden
	}

	return s.repo.GetLabelsByTaskID(ctx, taskID)
}
