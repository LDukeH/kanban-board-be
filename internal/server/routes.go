package server

import (
	"github.com/go-chi/chi/v5/middleware"

	"kanban-board-be/internal/hello"
)

func (s *Server) MountHandlers() {
    s.Router.Use(middleware.Logger)

    s.Router.Get("/", hello.HelloWorld)
    s.Router.Get("/health", HealthCheck)

    // s.Router.Post("/auth/login", LoginHandler)

    // s.Router.Get("/boards", GetBoards)
    // s.Router.Post("/boards", CreateBoard)
}