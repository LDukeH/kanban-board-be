package server

import (
	"github.com/go-chi/chi/v5"

	"kanban-board-be/internal/auth"
	"kanban-board-be/internal/user"
)

type Server struct{
	Router *chi.Mux
	//add other information like db here

	userHandler *user.Handler

	// auth related
	authHandler *auth.Handler
	authService *auth.Service
}

func CreateNewServer(
	userHandler *user.Handler, 
	authHandler *auth.Handler, 
	authService *auth.Service,
) *Server{
	s := &Server{
		userHandler: userHandler,
		authHandler: authHandler,
		authService: authService,
	}
	s.Router = chi.NewRouter()
	return s
}
