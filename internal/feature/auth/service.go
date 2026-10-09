package auth

import (
	"context"
	"errors"
	"log"

	"kanban-board-be/internal/apperrors"
	"kanban-board-be/internal/feature/user"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Service struct {
	userService *user.Service
	config      Config
	jwks        keyfunc.Keyfunc
}

func NewService(userService *user.Service, config Config, jwks keyfunc.Keyfunc) *Service {
	return &Service{
		userService: userService,
		config:      config,
		jwks:        jwks,
	}
}

func (s *Service) SyncUser(ctx context.Context, id uuid.UUID, email string) error {
	return s.userService.SyncExternalUser(ctx, id, email)
}

func (s *Service) LookupUser(ctx context.Context) (*user.User, error) {
	userID, err := GetUserIDCtx(ctx)
	if err != nil {
		return nil, err
	}

	return s.userService.GetUserByID(ctx, userID.String())
}

func (s *Service) DeleteMyAccount(ctx context.Context) error {
	userID, err := GetUserIDCtx(ctx)
	if err != nil {
		return err
	}

	return s.userService.DeleteUserByID(ctx, userID.String())
}

func (s *Service) ValidateToken(tokenString string) (string, error) {
	if s.jwks == nil {
		return "", errors.Join(apperrors.ErrUnauthorized, errors.New("Supabase JWT verifier is not configured"))
	}

	token, err := jwt.Parse(tokenString, s.jwks.Keyfunc)

	if err != nil {
		log.Printf("Failed to parse token: %v", err)
		return "", errors.Join(apperrors.ErrUnauthorized, err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok || !token.Valid {
		log.Printf("Invalid token claims: %v", claims)
		return "", errors.Join(apperrors.ErrUnauthorized, jwt.ErrSignatureInvalid)
	}

	if s.config.Issuer != "" {
		issuer, ok := claims["iss"].(string)
		if !ok || issuer != s.config.Issuer {
			return "", errors.Join(apperrors.ErrUnauthorized, jwt.ErrTokenInvalidIssuer)
		}
	}

	if role, ok := claims["role"].(string); !ok || role != "authenticated" {
		return "", errors.Join(apperrors.ErrUnauthorized, jwt.ErrTokenInvalidClaims)
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		log.Printf("Subject not found in token claims: %v", claims)
		return "", errors.Join(apperrors.ErrUnauthorized, jwt.ErrSignatureInvalid)
	}
	return userID, nil
}
