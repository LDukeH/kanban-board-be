package auth

import (
	"context"
	"log"
	"net/http"
)

func (s *Service) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the token from the cookie
		cookie, err := r.Cookie("access_token")

		if err != nil {
			if err == http.ErrNoCookie {
				http.Error(w, "Missing access token", http.StatusUnauthorized)
			} else {
				log.Printf("failed to read access token cookie: %v", err)
				http.Error(w, "Invalid request", http.StatusBadRequest)
			}
			return
		}

		if cookie == nil || cookie.Value == "" {
			http.Error(w, "Missing access token", http.StatusUnauthorized)
			return
		}

		tokenString := cookie.Value

		// Validate through the service
		userID, err := ValidateToken(tokenString, s.config.SecretKey)

		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			log.Printf("Invalid token: %v", err)
			return
		}

		// Attach the user ID to the request context, userIDKey is from model
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
