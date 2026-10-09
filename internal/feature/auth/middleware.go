package auth

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func (s *Service) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := strings.TrimSpace(r.Header.Get("Authorization"))
		scheme, tokenString, ok := strings.Cut(header, " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(tokenString) == "" {
			http.Error(w, "Missing bearer token", http.StatusUnauthorized)
			return
		}

		userID, err := s.ValidateToken(strings.TrimSpace(tokenString))

		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			log.Printf("Invalid token: %v", err)
			return
		}

		token, err := jwt.Parse(strings.TrimSpace(tokenString), s.jwks.Keyfunc)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		if email, ok := claims["email"].(string); ok {
			ctx = context.WithValue(ctx, UserEmailKey, email)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
