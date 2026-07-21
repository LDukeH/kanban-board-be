package server

import (

	"github.com/go-chi/chi/v5"
)

type Server struct{
	Router *chi.Mux

	//add other information like db here
}

func CreateNewServer() *Server{
	s := &Server{}
	s.Router = chi.NewRouter()
	return s
}
