package main

import (
	"log"

	"github.com/joho/godotenv"

	"net/http"

	"os"

	"time"

	"kanban-board-be/internal/server"

	"kanban-board-be/internal/db"
	"kanban-board-be/internal/feature/auth"
	"kanban-board-be/internal/feature/board"
	"kanban-board-be/internal/feature/membership"
	"kanban-board-be/internal/feature/user"
	"kanban-board-be/internal/feature/workspace"
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

	authService := auth.NewService(userService, authConfig)
	authHandler := auth.NewHandler(authService)

	// Workspace injection
	workspaceRepo := workspace.NewRepository(pool)
	membershipRepo := membership.NewRepository(pool)
	workspaceService := workspace.NewService(workspaceRepo, membershipRepo)
	workspaceHandler := workspace.NewHandler(workspaceService)

	// Board injection
	boardRepo := board.NewRepository(pool)
	boardService := board.NewService(boardRepo, workspaceRepo)
	boardHandler := board.NewHandler(boardService)

	// Create server
	s := server.CreateNewServer(userHandler, authHandler, authService, workspaceHandler, workspaceService, boardHandler, boardService)
	s.MountHandlers()

	// Start HTTP server
	log.Fatal(http.ListenAndServe(":3000", s.Router))

	_ = userRepo
	_ = userService
	_ = userHandler
	_ = authService
	_ = authHandler
}