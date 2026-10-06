package comment

import (
	"context"
	"errors"
	"strings"

	"kanban-board-be/internal/apperrors"
	"kanban-board-be/internal/feature/authz"
	"kanban-board-be/internal/feature/membership"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const maxCommentLength = 5000

type commentRepository interface {
	ListByTaskID(ctx context.Context, taskID uuid.UUID) ([]*CommentResponse, error)
	Create(ctx context.Context, comment *Comment) (*CommentResponse, error)
	GetByID(ctx context.Context, commentID uuid.UUID) (*Comment, error)
	Delete(ctx context.Context, commentID uuid.UUID) error
}

type Service struct {
	commentRepo commentRepository
	authz       *authz.Service
}

func NewService(commentRepo commentRepository, authz *authz.Service) *Service {
	return &Service{
		commentRepo: commentRepo,
		authz:       authz,
	}
}

func (s *Service) ListByTaskID(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) ([]*CommentResponse, error) {
	role, err := s.authz.GetTaskRole(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}

	if role != membership.Owner && role != membership.Admin && role != membership.Member {
		return nil, apperrors.ErrForbidden
	}

	return s.commentRepo.ListByTaskID(ctx, taskID)
}

func (s *Service) CreateComment(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, content string) (*CommentResponse, error) {
	content = strings.TrimSpace(content)
	if len(content) < 1 {
		return nil, apperrors.ErrValidation
	}
	if len(content) > maxCommentLength {
		return nil, apperrors.ErrValidation
	}

	role, err := s.authz.GetTaskRole(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}

	if role != membership.Owner && role != membership.Admin && role != membership.Member {
		return nil, apperrors.ErrForbidden
	}

	comment := &Comment{
		ID:       uuid.New(),
		TaskID:   taskID,
		AuthorID: userID,
		Content:  content,
	}

	return s.commentRepo.Create(ctx, comment)
}

func (s *Service) DeleteComment(ctx context.Context, userID uuid.UUID, commentID uuid.UUID) error {
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return err
	}

	role, err := s.authz.GetTaskRole(ctx, userID, comment.TaskID)
	if err != nil {
		return err
	}

	// The author can always delete their own comment; owner/admin can delete
	// another user's comment. Do not rely on the frontend hiding the button.
	if comment.AuthorID != userID && role != membership.Owner && role != membership.Admin {
		return apperrors.ErrForbidden
	}

	err = s.commentRepo.Delete(ctx, commentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return err
	}

	return nil
}
