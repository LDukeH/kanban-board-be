package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"kanban-board-be/internal/apperrors"
	"kanban-board-be/internal/feature/user"
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
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("ERROR decoding request body:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}

	var user user.User
	user.Email = req.Email
	user.Password = req.Password

	if _, err := h.authService.RegisterUser(r.Context(), user.Email, user.Password); err != nil {
		log.Printf("Failed to register user: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("ERROR decoding request body:", err)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}

	token, err := h.authService.LoginUser(r.Context(), req.Email, req.Password)
	if err != nil {
		log.Printf("Failed to login user: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	http.SetCookie(w, newAccessTokenCookie(token, int(h.authService.config.Expiration.Seconds())))

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear the cookie by setting its MaxAge to -1
	http.SetCookie(w, clearAccessTokenCookie())

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteMyAccountHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.authService.DeleteMyAccount(r.Context()); err != nil {
		log.Printf("Failed to delete account: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	// Account is gone: clear the access_token cookie (same as logout)
	http.SetCookie(w, clearAccessTokenCookie())

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetMeHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.authService.LookupUser(r.Context())
	if err != nil {
		log.Printf("Failed to get user: %v", err)
		apperrors.WriteError(w, err)
		return
	}

	getMeResponse := GetMeResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339Nano),
	}

	json.NewEncoder(w).Encode(getMeResponse)
}
