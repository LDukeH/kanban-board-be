package invitation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"kanban-board-be/internal/apperrors"
	"kanban-board-be/internal/feature/membership"
	"kanban-board-be/internal/feature/notification"
	"kanban-board-be/internal/feature/user"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type invitationRepository interface {
	BeginTx(ctx context.Context) (pgx.Tx, error)
	CreateTx(ctx context.Context, tx pgx.Tx, invitation *WorkspaceInvitation) error
	GetByIDForUpdate(ctx context.Context, tx pgx.Tx, invitationID uuid.UUID) (*WorkspaceInvitation, error)
	UpdateStatusTx(ctx context.Context, tx pgx.Tx, invitationID uuid.UUID, status Status) (time.Time, error)
	ListByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*InvitationResponse, error)
}

type membershipRepository interface {
	CreateTx(ctx context.Context, tx pgx.Tx, membership *membership.WorkspaceMembership) error
	GetByUserAndWorkspaceID(ctx context.Context, userID, workspaceID uuid.UUID) (*membership.WorkspaceMembership, error)
}

type workspaceNameProvider interface {
	GetWorkspaceName(ctx context.Context, workspaceID uuid.UUID) (string, error)
}

type Service struct {
	invitationRepo      invitationRepository
	membershipRepo      membershipRepository
	workspaceRepo       workspaceNameProvider
	notificationService *notification.Service
	userService         *user.Service
}

func NewService(
	invitationRepo invitationRepository,
	membershipRepo membershipRepository,
	workspaceRepo workspaceNameProvider,
	notificationService *notification.Service,
	userService *user.Service,
) *Service {
	return &Service{
		invitationRepo:      invitationRepo,
		membershipRepo:      membershipRepo,
		workspaceRepo:       workspaceRepo,
		notificationService: notificationService,
		userService:         userService,
	}
}

func (s *Service) Create(ctx context.Context, workspaceID, inviteeUserID, actorID uuid.UUID) (*WorkspaceInvitation, error) {
	// Only owners and admins may invite.
	actorRole, err := s.membershipRepo.GetByUserAndWorkspaceID(ctx, actorID, workspaceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrForbidden
		}
		return nil, err
	}

	if actorRole.Role != membership.Owner && actorRole.Role != membership.Admin {
		return nil, apperrors.ErrForbidden
	}

	// Only registered users can be invited.
	if _, err := s.userService.GetUserByID(ctx, inviteeUserID.String()); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to load invitee: %w", err)
	}

	// Existing workspace members cannot be invited.
	if _, err := s.membershipRepo.GetByUserAndWorkspaceID(ctx, inviteeUserID, workspaceID); err == nil {
		return nil, apperrors.ErrConflict
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	invitation := &WorkspaceInvitation{
		WorkspaceID:     workspaceID,
		InviteeUserID:   inviteeUserID,
		InvitedByUserID: actorID,
		Status:          StatusPending,
		ExpiresAt:       time.Now().Add(DefaultExpiry),
	}

	// The invitation and its notification must be all-or-nothing.
	tx, err := s.invitationRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.invitationRepo.CreateTx(ctx, tx, invitation); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, apperrors.ErrConflict
		}
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	workspaceName, err := s.workspaceRepo.GetWorkspaceName(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace name for notification: %w", err)
	}

	actor, err := s.userService.GetUserByID(ctx, actorID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get actor for notification: %w", err)
	}

	if err := s.notificationService.NotifyWorkspaceInvitedTx(ctx, tx, inviteeUserID, invitation.ID, workspaceID, workspaceName, actor.Email); err != nil {
		return nil, fmt.Errorf("failed to create workspace invitation notification: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit invitation creation: %w", err)
	}

	return invitation, nil
}

func (s *Service) List(ctx context.Context, workspaceID, userID uuid.UUID) ([]*InvitationResponse, error) {
	// Only workspace owners and admins may view the invitation list.
	mem, err := s.membershipRepo.GetByUserAndWorkspaceID(ctx, userID, workspaceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrForbidden
		}
		return nil, err
	}

	if mem.Role != membership.Owner && mem.Role != membership.Admin {
		return nil, apperrors.ErrForbidden
	}

	invitations, err := s.invitationRepo.ListByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	for _, invitation := range invitations {
		invitation.Status = effectiveStatus(invitation.Status, invitation.ExpiresAt, now)
	}

	return invitations, nil
}

func (s *Service) Accept(ctx context.Context, invitationID, userID uuid.UUID) (*WorkspaceInvitation, error) {
	tx, invitation, err := s.beginAndLock(ctx, invitationID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Only the invited user may accept.
	if invitation.InviteeUserID != userID {
		return nil, apperrors.ErrForbidden
	}

	if err := s.requirePending(ctx, tx, invitation); err != nil {
		return nil, err
	}

	updatedAt, err := s.invitationRepo.UpdateStatusTx(ctx, tx, invitation.ID, StatusAccepted)
	if err != nil {
		return nil, fmt.Errorf("failed to accept invitation: %w", err)
	}
	invitation.Status = StatusAccepted
	invitation.UpdatedAt = updatedAt

	newMembership := &membership.WorkspaceMembership{
		WorkspaceID: invitation.WorkspaceID,
		UserID:      invitation.InviteeUserID,
		Role:        membership.Member,
	}

	if err := s.membershipRepo.CreateTx(ctx, tx, newMembership); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// The invitee became a member some other way in the meantime.
			return nil, apperrors.ErrConflict
		}
		return nil, fmt.Errorf("failed to create workspace membership: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit invitation acceptance: %w", err)
	}

	return invitation, nil
}

func (s *Service) Decline(ctx context.Context, invitationID, userID uuid.UUID) (*WorkspaceInvitation, error) {
	tx, invitation, err := s.beginAndLock(ctx, invitationID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Only the invited user may decline.
	if invitation.InviteeUserID != userID {
		return nil, apperrors.ErrForbidden
	}

	if err := s.requirePending(ctx, tx, invitation); err != nil {
		return nil, err
	}

	updatedAt, err := s.invitationRepo.UpdateStatusTx(ctx, tx, invitation.ID, StatusDeclined)
	if err != nil {
		return nil, fmt.Errorf("failed to decline invitation: %w", err)
	}
	invitation.Status = StatusDeclined
	invitation.UpdatedAt = updatedAt

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit invitation decline: %w", err)
	}

	return invitation, nil
}

func (s *Service) Cancel(ctx context.Context, workspaceID, invitationID, userID uuid.UUID) (*WorkspaceInvitation, error) {
	tx, invitation, err := s.beginAndLock(ctx, invitationID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// The invitation must belong to the workspace named in the request.
	if invitation.WorkspaceID != workspaceID {
		return nil, apperrors.ErrNotFound
	}

	actorRole, err := s.membershipRepo.GetByUserAndWorkspaceID(ctx, userID, workspaceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrForbidden
		}
		return nil, err
	}

	if actorRole.Role != membership.Owner && actorRole.Role != membership.Admin {
		return nil, apperrors.ErrForbidden
	}

	if err := s.requirePending(ctx, tx, invitation); err != nil {
		return nil, err
	}

	updatedAt, err := s.invitationRepo.UpdateStatusTx(ctx, tx, invitation.ID, StatusCancelled)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel invitation: %w", err)
	}
	invitation.Status = StatusCancelled
	invitation.UpdatedAt = updatedAt

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit invitation cancellation: %w", err)
	}

	return invitation, nil
}

func (s *Service) beginAndLock(ctx context.Context, invitationID uuid.UUID) (pgx.Tx, *WorkspaceInvitation, error) {
	tx, err := s.invitationRepo.BeginTx(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	invitation, err := s.invitationRepo.GetByIDForUpdate(ctx, tx, invitationID)
	if err != nil {
		tx.Rollback(ctx)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, apperrors.ErrNotFound
		}
		return nil, nil, fmt.Errorf("failed to load invitation: %w", err)
	}

	return tx, invitation, nil
}

func (s *Service) requirePending(ctx context.Context, tx pgx.Tx, invitation *WorkspaceInvitation) error {
	if invitation.EffectiveStatus(time.Now()) == StatusPending {
		return nil
	}

	if invitation.Status == StatusPending {
		updatedAt, err := s.invitationRepo.UpdateStatusTx(ctx, tx, invitation.ID, StatusExpired)
		if err != nil {
			return fmt.Errorf("failed to expire invitation: %w", err)
		}
		invitation.Status = StatusExpired
		invitation.UpdatedAt = updatedAt

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit invitation expiry: %w", err)
		}
	}

	return apperrors.ErrConflict
}
