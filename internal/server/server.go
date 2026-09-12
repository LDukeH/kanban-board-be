package server

import (
	"github.com/go-chi/chi/v5"

	"kanban-board-be/internal/feature/auth"
	"kanban-board-be/internal/feature/board"
	"kanban-board-be/internal/feature/column"
	"kanban-board-be/internal/feature/label"
	"kanban-board-be/internal/feature/membership"
	"kanban-board-be/internal/feature/notification"
	"kanban-board-be/internal/feature/task"
	"kanban-board-be/internal/feature/user"
	"kanban-board-be/internal/feature/workspace"
)

type Server struct {
	Router *chi.Mux
	//add other information like db here

	userHandler *user.Handler

	// auth related
	authHandler *auth.Handler
	authService *auth.Service

	// workspaceHandler *workspace.Handler
	workspaceHandler *workspace.Handler
	workspaceService *workspace.Service

	// membershipHandler *membership.Handler
	membershipHandler *membership.Handler
	membershipService *membership.Service

	// boardHandler *board.Handler
	boardHandler *board.Handler
	boardService *board.Service

	// columnHandler *column.Handler
	columnHandler *column.Handler
	columnService *column.Service

	// taskHandler *task.Handler
	taskHandler *task.Handler
	taskService *task.Service

	// labelHandler *label.Handler
	labelHandler *label.Handler
	labelService *label.Service

	// notificationHandler *notification.Handler
	notificationHandler *notification.Handler
	notificationService *notification.Service
}

func CreateNewServer(
	userHandler *user.Handler,
	authHandler *auth.Handler,
	authService *auth.Service,
	workspaceHandler *workspace.Handler,
	workspaceService *workspace.Service,
	membershipHandler *membership.Handler,
	membershipService *membership.Service,
	boardHandler *board.Handler,
	boardService *board.Service,
	columnHandler *column.Handler,
	columnService *column.Service,
	taskHandler *task.Handler,
	taskService *task.Service,
	labelHandler *label.Handler,
	labelService *label.Service,
	notificationHandler *notification.Handler,
	notificationService *notification.Service,
) *Server {
	s := &Server{
		userHandler:         userHandler,
		authHandler:         authHandler,
		authService:         authService,
		workspaceHandler:    workspaceHandler,
		workspaceService:    workspaceService,
		membershipHandler:   membershipHandler,
		membershipService:   membershipService,
		boardHandler:        boardHandler,
		boardService:        boardService,
		columnHandler:       columnHandler,
		columnService:       columnService,
		taskHandler:         taskHandler,
		taskService:         taskService,
		labelHandler:        labelHandler,
		labelService:        labelService,
		notificationHandler: notificationHandler,
		notificationService: notificationService,
	}
	s.Router = chi.NewRouter()
	return s
}
