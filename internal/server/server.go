package server

import (
	"github.com/go-chi/chi/v5"

	"kanban-board-be/internal/feature/auth"
	"kanban-board-be/internal/feature/user"
	"kanban-board-be/internal/feature/workspace"
	"kanban-board-be/internal/feature/board"
)

type Server struct{
	Router *chi.Mux
	//add other information like db here

	userHandler *user.Handler

	// auth related
	authHandler *auth.Handler
	authService *auth.Service

	// workspaceHandler *workspace.Handler
	workspaceHandler *workspace.Handler
	workspaceService *workspace.Service

	// boardHandler *board.Handler
	boardHandler *board.Handler
	boardService *board.Service
}

func CreateNewServer(
	userHandler *user.Handler, 
	authHandler *auth.Handler, 
	authService *auth.Service,
	workspaceHandler *workspace.Handler,
	workspaceService *workspace.Service,
	boardHandler *board.Handler,
	boardService *board.Service,
) *Server{
	s := &Server{
		userHandler: userHandler,
		authHandler: authHandler,
		authService: authService,
		workspaceHandler: workspaceHandler,
		workspaceService: workspaceService,
		boardHandler: boardHandler,
		boardService: boardService,
	}
	s.Router = chi.NewRouter()
	return s
}
