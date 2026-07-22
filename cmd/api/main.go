package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"

	"net/http"

	"kanban-board-be/internal/server"

	"kanban-board-be/internal/db"

	"kanban-board-be/internal/user"

	"context"

	"github.com/google/uuid"

)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env")
	}

	pool, err := db.NewPostgresPool()
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	userRepo := user.NewRepository(pool)

	// Test repository
	user, err := userRepo.GetByID(
		context.Background(),
		uuid.MustParse("44a1168e-43ec-4f70-9b4c-349a6e80ae63"),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(user)

	// Create server
	s := server.CreateNewServer()
	s.MountHandlers()

	// Start HTTP server
	log.Fatal(http.ListenAndServe(":3000", s.Router))
}