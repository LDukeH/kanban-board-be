package server

import (
	"github.com/go-chi/chi/v5"
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

        s.Router.Route("/auth", func(r chi.Router) {
        r.Post("/register", s.authHandler.RegisterHandler)
        r.Post("/login", s.authHandler.LoginHandler)

        // Protected auth routes
        r.Group(func(r chi.Router) {
            r.Use(s.authService.AuthMiddleware)

            r.Post("/logout", s.authHandler.LogoutHandler)
            r.Get("/me", s.authHandler.GetMeHandler)
        })

        // Protected API
        s.Router.Group(func(r chi.Router) {
        r.Use(s.authService.AuthMiddleware)

        r.Get("/users/{id}", s.userHandler.GetUserByID)

        // Workspace routes
        r.Route("/workspaces", func(r chi.Router) {
            r.Post("/", s.workspaceHandler.CreateWorkspace)
            r.Get("/{id}", s.workspaceHandler.GetWorkspaceByID)
            r.Get("/", s.workspaceHandler.GetAllWorkspaces)

            // Board routes inside workspace
            r.Route("/{workspaceID}/boards", func(r chi.Router) {
                r.Post("/", s.boardHandler.CreateBoard)
                r.Get("/", s.boardHandler.GetBoardsByWorkspaceID)
        })
    })
        // Other protected routes can be added here
        // Board routes
        r.Route("/boards", func(r chi.Router) {
            r.Get("/{boardID}", s.boardHandler.GetBoardByID)
            // r.Put("/{id}", s.boardHandler.UpdateBoard)
            // r.Delete("/{id}", s.boardHandler.DeleteBoard)

            // Column routes inside board
            r.Route("/{boardID}/columns", func(r chi.Router) {
                r.Post("/", s.columnHandler.CreateColumn)
                r.Get("/", s.columnHandler.GetColumnsByBoardID)
            })
        })
    })
    })
}

    func helloHandler(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Hello, World!"))
    }