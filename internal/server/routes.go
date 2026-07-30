package server

import (
	"github.com/go-chi/chi/v5/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func (s *Server) MountHandlers() {
	s.Router.Use(middleware.Logger)

	s.Router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public routes


	s.Router.Post("/auth/register", s.authHandler.RegisterHandler)
	s.Router.Post("/auth/login", s.authHandler.LoginHandler)

	// Protected auth routes

	s.Router.Group(func(r chi.Router) {
		r.Use(s.authService.AuthMiddleware)

		r.Post("/auth/logout", s.authHandler.LogoutHandler)
		r.Get("/auth/me", s.authHandler.GetMeHandler)
	})

	// Protected API routes
	s.Router.Group(func(r chi.Router) {
		r.Use(s.authService.AuthMiddleware)

		// Users
		r.Get("/users/{id}", s.userHandler.GetUserByID)

		// Workspaces
		r.Route("/workspaces", func(r chi.Router) {
			r.Post("/", s.workspaceHandler.CreateWorkspace)
			r.Get("/", s.workspaceHandler.GetAllWorkspaces)
			r.Get("/{id}", s.workspaceHandler.GetWorkspaceByID)
			r.Patch("/{id}", s.workspaceHandler.UpdateWorkspace)
			r.Delete("/{id}", s.workspaceHandler.DeleteWorkspace)

			// Boards inside workspace
			r.Route("/{workspaceID}/boards", func(r chi.Router) {
				r.Post("/", s.boardHandler.CreateBoard)
				r.Get("/", s.boardHandler.GetBoardsByWorkspaceID)
                r.Patch("/{id}", s.boardHandler.UpdateBoard)
                r.Delete("/{id}", s.boardHandler.DeleteBoard)
			})
		})

		// Boards
		r.Route("/boards", func(r chi.Router) {
			r.Get("/{id}", s.boardHandler.GetBoardByID)

			// Columns inside board
			r.Route("/{boardID}/columns", func(r chi.Router) {
				r.Post("/", s.columnHandler.CreateColumn)
				r.Get("/", s.columnHandler.GetColumnsByBoardID)
			})
		})
	})
}