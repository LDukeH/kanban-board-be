package server

import (
	"github.com/go-chi/chi/v5"

	"kanban-board-be/internal/user"
	"kanban-board-be/internal/auth"
)

type Server struct{
	Router *chi.Mux
	//add other information like db here

	userHandler *user.Handler
	authHandler *auth.Handler
}

func CreateNewServer(userHandler *user.Handler, authHandler *auth.Handler) *Server{
	s := &Server{
		userHandler: userHandler,
		authHandler: authHandler,
	}
	s.Router = chi.NewRouter()
	return s
}
