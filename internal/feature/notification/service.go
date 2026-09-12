package notification

import (
	"context"
	"encoding/json"
	"fmt"

	"kanban-board-be/internal/apperrors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type notificationRepository interface {
	Create(ctx context.Context, notification *Notification) error
	CreateTx(ctx context.Context, tx pgx.Tx, notification *Notification) error
	ListByUserID(ctx context.Context, userID uuid.UUID, limit int, unreadOnly bool) ([]*Notification, error)
	CountUnreadByUserID(ctx context.Context, userID uuid.UUID) (int, error)
	MarkAsRead(ctx context.Context, notificationID uuid.UUID, userID uuid.UUID) (bool, error)
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
}

type Service struct {
	notificationRepo notificationRepository
}

func NewService(notificationRepo notificationRepository) *Service {
	return &Service{
		notificationRepo: notificationRepo,
	}
}

func (s *Service) List(ctx context.Context, userID uuid.UUID, limit int, unreadOnly bool) ([]*Notification, error) {
	return s.notificationRepo.ListByUserID(ctx, userID, limit, unreadOnly)
}

func (s *Service) UnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.notificationRepo.CountUnreadByUserID(ctx, userID)
}

// Another user's notification behaves as not found; marking is idempotent.
func (s *Service) MarkAsRead(ctx context.Context, notificationID uuid.UUID, userID uuid.UUID) error {
	updated, err := s.notificationRepo.MarkAsRead(ctx, notificationID, userID)
	if err != nil {
		return err
	}

	if !updated {
		return apperrors.ErrNotFound
	}

	return nil
}

func (s *Service) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return s.notificationRepo.MarkAllAsRead(ctx, userID)
}

// Creates a task_assigned notification in the caller's transaction.
func (s *Service) NotifyTaskAssignedTx(
	ctx context.Context,
	tx pgx.Tx,
	recipientUserID uuid.UUID,
	taskID uuid.UUID,
	taskTitle string,
	actorEmail string,
) error {
	payload, err := json.Marshal(TaskAssignedPayload{
		TaskID:     taskID,
		TaskTitle:  taskTitle,
		ActorEmail: actorEmail,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal task_assigned payload: %w", err)
	}

	notification := &Notification{
		UserID:  recipientUserID,
		Type:    TypeTaskAssigned,
		Payload: payload,
	}

	return s.notificationRepo.CreateTx(ctx, tx, notification)
}

// Creates a workspace_invited notification in the caller's transaction.
func (s *Service) NotifyWorkspaceInvitedTx(
	ctx context.Context,
	tx pgx.Tx,
	recipientUserID uuid.UUID,
	workspaceID uuid.UUID,
	workspaceName string,
	actorEmail string,
) error {
	payload, err := json.Marshal(WorkspaceInvitedPayload{
		WorkspaceID:   workspaceID,
		WorkspaceName: workspaceName,
		ActorEmail:    actorEmail,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal workspace_invited payload: %w", err)
	}

	notification := &Notification{
		UserID:  recipientUserID,
		Type:    TypeWorkspaceInvited,
		Payload: payload,
	}

	return s.notificationRepo.CreateTx(ctx, tx, notification)
}
