package auth

import (
	"net/http"

	"kanban-board-be/internal/feature/user"

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
		log.Printf("Failed to register user: %v", err)
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK) // 200 OK

	json.NewEncoder(w).Encode(map[string]string{
		"message": "User registered successfully",
	})
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// for now just decode the email and password from the request body and log them
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
		http.Error(w, "Failed to login user", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
    Name:     "access_token",
    Value:    token, // Your JWT string
    Path:     "/",
    HttpOnly: true,
    Secure:   false,               // Set to true in production (HTTPS)
    SameSite: http.SameSiteLaxMode,
})

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear the cookie by setting its MaxAge to -1
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production (HTTPS)
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h* Handler) GetMeHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get user from database
	user, err := h.authService.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to get user by ID: %v", err)
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(user)

}
