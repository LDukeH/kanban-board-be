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
	// Extract the user ID from the URL path
	id := chi.URLParam(r, "id")

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