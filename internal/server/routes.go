package server

import (
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) MountHandlers() {
	s.Router.Use(middleware.Logger)

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	s.Router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{frontendURL},
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
		r.Delete("/auth/me", s.authHandler.DeleteMyAccountHandler)
	})

	// Protected API routes
	s.Router.Group(func(r chi.Router) {
		r.Use(s.authService.AuthMiddleware)

		// Users
		r.Get("/users/{userID}", s.userHandler.GetUserByID)
		r.Get("/users/search", s.userHandler.SearchUsersByEmail)

		// Workspaces
		r.Route("/workspaces", func(r chi.Router) {
			r.Post("/", s.workspaceHandler.CreateWorkspace)
			r.Get("/", s.workspaceHandler.GetAllWorkspaces)
			r.Get("/{workspaceID}", s.workspaceHandler.GetWorkspaceByID)
			r.Patch("/{workspaceID}", s.workspaceHandler.UpdateWorkspace)
			r.Delete("/{workspaceID}", s.workspaceHandler.DeleteWorkspace)
			r.Post("/{workspaceID}/transfer-ownership", s.workspaceHandler.TransferOwnership)
			r.Delete("/{workspaceID}/memberships", s.membershipHandler.LeaveWorkspace)

			// Boards inside workspace
			r.Route("/{workspaceID}/boards", func(r chi.Router) {
				r.Post("/", s.boardHandler.CreateBoard)
				r.Get("/", s.boardHandler.GetBoardsByWorkspaceID)
			})

			// Membership inside workspace
			r.Route("/{workspaceID}/members", func(r chi.Router) {
				r.Get("/", s.membershipHandler.GetMembershipsByWorkspaceID)
				r.Patch("/{membershipID}", s.membershipHandler.ChangeMemberRole)
				r.Delete("/{membershipID}", s.membershipHandler.RemoveMember)
			})

			// Workspace invitations
			r.Route("/{workspaceID}/invitations", func(r chi.Router) {
				r.Post("/", s.invitationHandler.CreateInvitation)
				r.Get("/", s.invitationHandler.GetInvitationsByWorkspaceID)
				r.Post("/{invitationID}/cancel", s.invitationHandler.CancelInvitation)
			})
		})

		// Invitations (invitee actions)
		r.Route("/invitations", func(r chi.Router) {
			r.Post("/{invitationID}/accept", s.invitationHandler.AcceptInvitation)
			r.Post("/{invitationID}/decline", s.invitationHandler.DeclineInvitation)
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

			// Users assigned to task
			r.Route("/{taskID}/users", func(r chi.Router) {
				r.Get("/", s.taskHandler.GetUsersByTaskID)
				r.Post("/", s.taskHandler.AssignUserToTask)
				r.Delete("/", s.taskHandler.RemoveUserFromTask)
			})

			// Comments attached to task
			r.Route("/{taskID}/comments", func(r chi.Router) {
				r.Get("/", s.commentHandler.ListComments)
				r.Post("/", s.commentHandler.CreateComment)
			})
		})

		// Comments
		r.Route("/comments", func(r chi.Router) {
			r.Delete("/{commentID}", s.commentHandler.DeleteComment)
		})

		// Notifications
		r.Route("/notifications", func(r chi.Router) {
			r.Get("/", s.notificationHandler.ListNotifications)
			r.Get("/unread-count", s.notificationHandler.GetUnreadCount)
			r.Post("/{notificationID}/read", s.notificationHandler.MarkNotificationAsRead)
			r.Post("/read-all", s.notificationHandler.MarkAllNotificationsAsRead)
		})
	})
}
