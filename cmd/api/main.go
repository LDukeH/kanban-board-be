package main

import (
	"log"

	"github.com/joho/godotenv"

	"net/http"

	"os"

	"time"

	"kanban-board-be/internal/server"

	"kanban-board-be/internal/db"
	"kanban-board-be/internal/feature/user"
	"kanban-board-be/internal/feature/auth"
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

	// Create server
	s := server.CreateNewServer(userHandler, authHandler, authService)
	s.MountHandlers()

	// Start HTTP server
	log.Fatal(http.ListenAndServe(":3000", s.Router))

	_ = userRepo
	_ = userService
	_ = userHandler
	_ = authService
	_ = authHandler
}