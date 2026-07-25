package server

import (
	"github.com/go-chi/chi/v5/middleware"

	"net/http"

	"github.com/go-chi/cors"
)

    func (s *Server) MountHandlers() {
        s.Router.Use(middleware.Logger)
        s.Router.Use(cors.Handler(cors.Options{
            AllowedOrigins:   []string{"http://localhost:5173"},
            AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
            AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
            AllowCredentials: true,
            MaxAge:           300, // Maximum value not ignored by any of major browsers
        }))
        s.Router.Get("/",helloHandler)

        s.Router.Get("/users/{id}", s.userHandler.GetUserByID)

        // Auth
        s.Router.Post("/auth/register", s.authHandler.RegisterHandler)
        s.Router.Post("/auth/login", s.authHandler.LoginHandler)
        s.Router.Post("/auth/logout", s.authHandler.LogoutHandler)


        // Testing protected route with auth middleware
        s.Router.With(s.authService.AuthMiddleware).Get("/auth/me", s.authHandler.GetMeHandler)
    }

    func helloHandler(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Hello, World!"))
    }