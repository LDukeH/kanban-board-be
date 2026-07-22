package server

import (
	"github.com/go-chi/chi/v5/middleware"

	"net/http"
)

func (s *Server) MountHandlers() {
    s.Router.Use(middleware.Logger)
    s.Router.Get("/",helloHandler)

    s.Router.Get("/users/{id}", s.userHandler.GetUserByID)

    // Auth
    s.Router.Post("/auth/register", s.authHandler.RegisterHandler)
    s.Router.Post("/auth/login", s.authHandler.LoginHandler)


    // s.Router.Get("/boards", GetBoards)
    // s.Router.Post("/boards", CreateBoard)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Hello, World!"))
}