package membership

import (
	"context"
	"errors"

	"kanban-board-be/internal/apperrors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type membershipRepository interface {
	CreateTx(ctx context.Context, tx pgx.Tx, membership *WorkspaceMembership) error
	ChangeMemberRole(ctx context.Context, membership *WorkspaceMembership) error
	RemoveMember(ctx context.Context, workspaceMembership *WorkspaceMembership) error
	GetByUserAndWorkspaceID(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID) (*WorkspaceMembership, error)
	GetByID(ctx context.Context, membershipID uuid.UUID) (*WorkspaceMembership, error)
	GetMemberListByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*WorkspaceMemberResponse, error)
}

type Service struct {
	membershipRepo membershipRepository
}

func NewService(membershipRepo membershipRepository) *Service {
	return &Service{
		membershipRepo: membershipRepo,
	}
}

func (s *Service) ChangeMemberRole(ctx context.Context, membership *WorkspaceMembership, userID uuid.UUID) error {
	// Check if the user is owner
	memberRole, err := s.membershipRepo.GetByUserAndWorkspaceID(ctx, userID, membership.WorkspaceID)
	if err != nil {
		return err
	}

	role := memberRole.Role

	if role != Owner {
		return apperrors.ErrForbidden
	}

	targetMembership, err := s.membershipRepo.GetByID(ctx, membership.ID)
	if err != nil {
		return err
	}

	if targetMembership.WorkspaceID != membership.WorkspaceID {
		return apperrors.ErrNotFound
	}

	if targetMembership.UserID == userID {
		return apperrors.ErrForbidden
	}

	if targetMembership.Role == Owner {
		return apperrors.ErrForbidden
	}

	return s.membershipRepo.ChangeMemberRole(ctx, membership)
}

func (s *Service) RemoveMember(ctx context.Context, workspaceMembership *WorkspaceMembership, userID uuid.UUID) error {
	// Check if the user is owner
	memberRole, err := s.membershipRepo.GetByUserAndWorkspaceID(ctx, userID, workspaceMembership.WorkspaceID)
	if err != nil {
		return err
	}

	role := memberRole.Role

	if role != Owner {
		return apperrors.ErrForbidden
	}

	targetMembership, err := s.membershipRepo.GetByID(ctx, workspaceMembership.ID)
	if err != nil {
		return err
	}

	if targetMembership.WorkspaceID != workspaceMembership.WorkspaceID {
		return apperrors.ErrNotFound
	}

	if targetMembership.UserID == userID {
		return apperrors.ErrForbidden
	}

	if targetMembership.Role == Owner {
		return apperrors.ErrForbidden
	}

	return s.membershipRepo.RemoveMember(ctx, targetMembership)
}

func (s *Service) GetMembershipByID(ctx context.Context, membershipID uuid.UUID) (*WorkspaceMembership, error) {
	return s.membershipRepo.GetByID(ctx, membershipID)
}

func (s *Service) GetMemberListByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID) ([]*WorkspaceMemberResponse, error) {
	// Check if the user is a member of the workspace
	memberRole, err := s.membershipRepo.GetByUserAndWorkspaceID(ctx, userID, workspaceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrForbidden
		}
		return nil, err
	}

	role := memberRole.Role
	if role != Owner && role != Admin && role != Member {
		return nil, apperrors.ErrForbidden
	}

	memberships, err := s.membershipRepo.GetMemberListByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	return memberships, nil
}

func (s *Service) LeaveWorkspace(ctx context.Context, workspaceMembership *WorkspaceMembership, userID uuid.UUID) error {
	// Check if the user is a member of the workspace
	membership, err := s.membershipRepo.GetByUserAndWorkspaceID(ctx, userID, workspaceMembership.WorkspaceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrForbidden
		}
		return err
	}

	role := membership.Role
	if role != Owner && role != Admin && role != Member {
		return apperrors.ErrForbidden
	}

	if membership.WorkspaceID != workspaceMembership.WorkspaceID {
		return apperrors.ErrNotFound
	}

	if membership.UserID != userID {
		return apperrors.ErrForbidden
	}

	// Prevent the last owner from leaving the workspace
	if membership.Role == Owner {
		return apperrors.ErrForbidden
	}

	return s.membershipRepo.RemoveMember(ctx, membership)
}
