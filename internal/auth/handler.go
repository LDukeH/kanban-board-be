package auth

import (
	"net/http"

	"kanban-board-be/internal/user"

	"encoding/json"

	"log"
)

type Handler struct {
	authService *Service
}

func NewHandler(authService *Service) *Handler {
	return &Handler{
		authService: authService,
	}
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user user.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Println("ERROR decoding request body:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if user.Email ==""{
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	if(user.Password==""){
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}

	if _, err := h.authService.RegisterUser(user.Email, user.Password); err != nil {
		log.Fatalf("Failed to register user: %v", err)
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK) // 200 OK

	json.NewEncoder(w).Encode(map[string]string{
		"message": "User registered successfully",
	})

}
