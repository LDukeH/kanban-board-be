package server

import (
	"github.com/go-chi/chi/v5/middleware"


)

func (s *Server) MountHandlers() {
    s.Router.Use(middleware.Logger)


    // s.Router.Post("/auth/login", LoginHandler)

    // s.Router.Get("/boards", GetBoards)
    // s.Router.Post("/boards", CreateBoard)
}