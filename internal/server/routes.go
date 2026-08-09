package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) MountHandlers() {
	s.Router.Use(middleware.Logger)

	s.Router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
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
			r.Get("/{workspaceID}", s.workspaceHandler.GetWorkspaceByID)
			r.Patch("/{workspaceID}", s.workspaceHandler.UpdateWorkspace)
			r.Delete("/{workspaceID}", s.workspaceHandler.DeleteWorkspace)

			// Boards inside workspace
			r.Route("/{workspaceID}/boards", func(r chi.Router) {
				r.Post("/", s.boardHandler.CreateBoard)
				r.Get("/", s.boardHandler.GetBoardsByWorkspaceID)
			})

			// Membership inside workspace
			r.Route("/{workspaceID}/members", func(r chi.Router) {
				r.Post("/", s.membershipHandler.CreateMembership)
				r.Get("/", s.membershipHandler.GetMembershipsByWorkspaceID)
				r.Patch("/{id}", s.membershipHandler.ChangeMemberRole)
				r.Delete("/{id}", s.membershipHandler.RemoveMember)
			})
		})

		// Boards
		r.Route("/boards", func(r chi.Router) {
			r.Get("/{boardID}", s.boardHandler.GetBoardByID)
			r.Patch("/{boardID}", s.boardHandler.UpdateBoard)
			r.Delete("/{boardID}", s.boardHandler.DeleteBoard)

			// Columns inside board
			r.Route("/{boardID}/columns", func(r chi.Router) {
				r.Post("/", s.columnHandler.CreateColumn)
				r.Get("/", s.columnHandler.GetColumnsByBoardID)
			})

			// Labels inside board
			r.Route("/{boardID}/labels", func(r chi.Router) {
				r.Post("/", s.labelHandler.CreateLabel)
				r.Get("/", s.labelHandler.GetLabelsByBoardID)
			})
		})

		// Columns
		r.Route("/columns", func(r chi.Router) {
			r.Patch("/{columnID}", s.columnHandler.UpdateColumn)
			r.Patch("/{columnID}/position", s.columnHandler.UpdateColumnPosition)
			r.Delete("/{columnID}", s.columnHandler.DeleteColumn)

			// Tasks inside column
			r.Route("/{columnID}/tasks", func(r chi.Router) {
				r.Post("/", s.taskHandler.CreateTask)
				r.Get("/", s.taskHandler.GetTasksByColumnID)
			})
		})

		// Labels
		r.Route("/labels", func(r chi.Router) {
			r.Get("/{labelID}", s.labelHandler.GetLabelByID)
			r.Patch("/{labelID}", s.labelHandler.UpdateLabel)
			r.Delete("/{labelID}", s.labelHandler.DeleteLabel)
		})

		// Tasks
		r.Route("/tasks", func(r chi.Router) {
		r.Patch("/{taskID}", s.taskHandler.UpdateTask)
		r.Patch("/{taskID}/position", s.taskHandler.UpdateTaskPosition)
		r.Delete("/{taskID}", s.taskHandler.DeleteTask)

		// Labels attached to task
		r.Route("/{taskID}/labels", func(r chi.Router) {
			r.Get("/", s.taskHandler.GetLabelsByTaskID)
			r.Post("/", s.taskHandler.AddLabelToTask)
			r.Delete("/{labelID}", s.taskHandler.RemoveLabelFromTask)
		})
	})
	})
}