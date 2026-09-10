package user

import (
	"encoding/json"
	"log"
	"net/http"

	"kanban-board-be/internal/apperrors"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "userID")

	user, err := h.service.GetUserByID(r.Context(), id)
	if err != nil {
		log.Printf("ERROR fetching user: %v\n", err)
		apperrors.WriteError(w, err)
		return
	}

	res := UserResponse{
		ID:    user.ID,
		Email: user.Email,
	}

	json.NewEncoder(w).Encode(res)
}

func (h *Handler) SearchUsersByEmail(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")

	users, err := h.service.SearchUsersByEmail(r.Context(), email)
	if err != nil {
		log.Printf("ERROR fetching user: %v\n", err)
		apperrors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
