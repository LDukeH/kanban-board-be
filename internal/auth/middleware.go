package auth

import (
	"net/http"

	"log"

	"context"

	"strings"
)

func (s *Service) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the token from the Authorization header
		authHeader := r.Header.Get("Authorization")

		if(authHeader == ""){
			http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
			log.Println("Authorization header is missing")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			log.Println("Invalid Authorization header format")
			return
		}

		tokenString := parts[1]

		// Validate through the service
		userID, err := ValidateToken(tokenString,s.config.SecretKey)

		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			log.Printf("Invalid token: %v", err)
			return
		}

		// Attach the user ID to the request context, userIDKey is from model
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}