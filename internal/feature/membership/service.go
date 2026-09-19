package membership

import (
	"context"
	"errors"
	"fmt"

	"kanban-board-be/internal/apperrors"
	"kanban-board-be/internal/feature/notification"
	"kanban-board-be/internal/feature/user"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type membershipRepository interface {
	BeginTx(ctx context.Context) (pgx.Tx, error)
	Create(ctx context.Context, membership *WorkspaceMembership) error
	CreateTx(ctx context.Context, tx pgx.Tx, membership *WorkspaceMembership) error
	ChangeMemberRole(ctx context.Context, membership *WorkspaceMembership) error
	RemoveMember(ctx context.Context, workspaceMembership *WorkspaceMembership) error
	GetByUserAndWorkspaceID(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID) (*WorkspaceMembership, error)
	GetByID(ctx context.Context, membershipID uuid.UUID) (*WorkspaceMembership, error)
	GetMemberListByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*WorkspaceMemberResponse, error)
}

// workspaceNameProvider loads a workspace's name. It is satisfied by the workspace repository.
type workspaceNameProvider interface {
	GetWorkspaceName(ctx context.Context, workspaceID uuid.UUID) (string, error)
}

type Service struct {
	membershipRepo      membershipRepository
	workspaceRepo       workspaceNameProvider
	notificationService *notification.Service
	userService         *user.Service
}

func NewService(
	membershipRepo membershipRepository,
	notificationService *notification.Service,
	userService *user.Service,
	workspaceRepo workspaceNameProvider,
) *Service {
	return &Service{
		membershipRepo:      membershipRepo,
		workspaceRepo:       workspaceRepo,
		notificationService: notificationService,
		userService:         userService,
	}
}

func (s *Service) CreateMembership(ctx context.Context, membership *WorkspaceMembership, userID uuid.UUID) error {
	// This is adding a new user to the workspace, so we need to check if the user is admin or owner
	// Check if the user is admin or owner
	memberRole, err := s.membershipRepo.GetByUserAndWorkspaceID(ctx, userID, membership.WorkspaceID)
	if err != nil {
		return err
	}

	role := memberRole.Role
	if role != Owner && role != Admin {
		return apperrors.ErrForbidden
	}

	// The membership and its notification must be all-or-nothing.
	tx, err := s.membershipRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.membershipRepo.CreateTx(ctx, tx, membership); err != nil {
		// Re-inviting an existing member violates the (user_id, workspace_id)
		// unique constraint: report it as a conflict, not a 500.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperrors.ErrConflict
		}
		return fmt.Errorf("failed to create membership: %w", err)
	}

	// Adding yourself to a workspace does not create a notification.
	if membership.UserID != userID {
		workspaceName, err := s.workspaceRepo.GetWorkspaceName(ctx, membership.WorkspaceID)
		if err != nil {
			return fmt.Errorf("failed to get workspace name for notification: %w", err)
		}

		actor, err := s.userService.GetUserByID(ctx, userID.String())
		if err != nil {
			return fmt.Errorf("failed to get actor for notification: %w", err)
		}

		if err := s.notificationService.NotifyWorkspaceInvitedTx(ctx, tx, membership.UserID, membership.WorkspaceID, workspaceName, actor.Email); err != nil {
			return fmt.Errorf("failed to create workspace invitation notification: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit membership creation: %w", err)
	}

	return nil
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

	// Pass the fully-loaded membership (which carries the real UserID) so the
	// repository can also remove the member's task assignments. The incoming
	// `workspaceMembership` from the handler only has ID + WorkspaceID.
	return s.membershipRepo.RemoveMember(ctx, targetMembership)
}

// GetMembershipByID returns a single membership by its ID, used to build
// update responses (e.g. after a role change).
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
