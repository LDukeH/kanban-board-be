package main

import (
	"net/http"

	"kanban-board-be/internal/server"
)

func main() {
	s := server.CreateNewServer()
	s.MountHandlers()
	http.ListenAndServe(":3000",s.Router);
}
