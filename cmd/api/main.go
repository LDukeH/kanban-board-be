package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"kanban-board-be/internal/db"
	"kanban-board-be/internal/feature/auth"
	"kanban-board-be/internal/feature/authz"
	"kanban-board-be/internal/feature/board"
	"kanban-board-be/internal/feature/column"
	"kanban-board-be/internal/feature/comment"
	"kanban-board-be/internal/feature/invitation"
	"kanban-board-be/internal/feature/label"
	"kanban-board-be/internal/feature/membership"
	"kanban-board-be/internal/feature/notification"
	"kanban-board-be/internal/feature/task"
	"kanban-board-be/internal/feature/user"
	"kanban-board-be/internal/feature/workspace"
	"kanban-board-be/internal/server"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	pool, err := db.NewPostgresPool(context.Background())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// User injection
	userRepo := user.NewRepository(pool)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	// Auth injection

	supabaseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("SUPABASE_URL")), "/")
	if supabaseURL == "" {
		log.Fatal("SUPABASE_URL is not set")
	}
	jwks, err := keyfunc.NewDefault([]string{supabaseURL + "/auth/v1/.well-known/jwks.json"})
	if err != nil {
		log.Fatalf("Failed to initialize Supabase JWT verifier: %v", err)
	}
	authConfig := auth.Config{
		JWKSURL: supabaseURL + "/auth/v1/.well-known/jwks.json",
		Issuer:  supabaseURL + "/auth/v1",
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
	authzService := authz.NewService(membershipRepo)
	authService := auth.NewService(userService, authConfig, jwks)
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

	membershipService := membership.NewService(membershipRepo)
	membershipHandler := membership.NewHandler(membershipService, userService)

	// Invitation
	invitationRepo := invitation.NewRepository(pool)
	invitationService := invitation.NewService(invitationRepo, membershipRepo, workspaceRepo, notificationService, userService)
	invitationHandler := invitation.NewHandler(invitationService, userService)

	// Comment
	commentService := comment.NewService(commentRepo, authzService)
	commentHandler := comment.NewHandler(commentService)

	// Create server
	s := server.CreateNewServer(
		userHandler, authHandler, authService,
		workspaceHandler, workspaceService,
		membershipHandler, membershipService,
		invitationHandler, invitationService,
		boardHandler, boardService,
		columnHandler, columnService,
		taskHandler, taskService,
		labelHandler, labelService,
		notificationHandler, notificationService,
		commentHandler, commentService)
	s.MountHandlers()

	// Start HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           s.Router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
