package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"kanban-board-be/internal/db"
	"kanban-board-be/internal/feature/auth"
	"kanban-board-be/internal/feature/authz"
	"kanban-board-be/internal/feature/board"
	"kanban-board-be/internal/feature/column"
	"kanban-board-be/internal/feature/comment"
	"kanban-board-be/internal/feature/label"
	"kanban-board-be/internal/feature/membership"
	"kanban-board-be/internal/feature/notification"
	"kanban-board-be/internal/feature/task"
	"kanban-board-be/internal/feature/user"
	"kanban-board-be/internal/feature/workspace"
	"kanban-board-be/internal/server"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	pool, err := db.NewPostgresPool()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// User injection
	userRepo := user.NewRepository(pool)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	// Auth injection

	authConfig := auth.Config{
		SecretKey:  os.Getenv("JWT_SECRET"),
		Expiration: 24 * time.Hour,
	}

	// Repo
	workspaceRepo := workspace.NewRepository(pool)
	membershipRepo := membership.NewRepository(pool)
	boardRepo := board.NewRepository(pool)
	columnRepo := column.NewRepository(pool)
	taskRepo := task.NewRepository(pool)
	labelRepo := label.NewRepository(pool)
	notificationRepo := notification.NewRepository(pool)
	commentRepo := comment.NewRepository(pool)
	// Authorization injection
	authzService := authz.NewService(membershipRepo, boardRepo, columnRepo, taskRepo, labelRepo)
	authService := auth.NewService(userService, authConfig)
	authHandler := auth.NewHandler(authService)

	// Notification
	notificationService := notification.NewService(notificationRepo)
	notificationHandler := notification.NewHandler(notificationService)

	// Board
	boardService := board.NewService(boardRepo, authzService)
	boardHandler := board.NewHandler(boardService)

	// Column
	columnService := column.NewService(boardRepo, columnRepo, authzService)
	columnHandler := column.NewHandler(columnService)

	// Label
	labelService := label.NewService(labelRepo, authzService)
	labelHandler := label.NewHandler(labelService)

	// Task
	taskService := task.NewService(taskRepo, authzService, notificationService, userService)
	taskHandler := task.NewHandler(taskService, labelService, userService)

	// Workspace injection
	workspaceService := workspace.NewService(workspaceRepo, membershipRepo, authzService)
	workspaceHandler := workspace.NewHandler(workspaceService)

	membershipService := membership.NewService(membershipRepo, notificationService, userService, workspaceRepo)
	membershipHandler := membership.NewHandler(membershipService, userService)

	// Comment
	commentService := comment.NewService(commentRepo, authzService)
	commentHandler := comment.NewHandler(commentService)

	// Create server
	s := server.CreateNewServer(
		userHandler, authHandler, authService,
		workspaceHandler, workspaceService,
		membershipHandler, membershipService,
		boardHandler, boardService,
		columnHandler, columnService,
		taskHandler, taskService,
		labelHandler, labelService,
		notificationHandler, notificationService,
		commentHandler, commentService)
	s.MountHandlers()

	// Start HTTP server
	log.Fatal(http.ListenAndServe(":3000", s.Router))

}
