package user

import (
	"net/http"

	"encoding/json"

	"log"

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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
}