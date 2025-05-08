package users

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type UserHandler struct {
	userService *UserService
}

func NewUserHandler(userService *UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetAllUsers(r.Context())
	if err != nil {
		log.Printf("handler: error fetching users: %v", err)

		switch {
		case errors.Is(err, ErrNoUsersFound):
			http.Error(w, "No users found", http.StatusNotFound)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		log.Printf("handler: error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *UserHandler) GetUserById(w http.ResponseWriter, r *http.Request) {
	var userId uuid.UUID
	userIdStr := r.URL.Query().Get("id") // Assuming the URL is /users?id=<uuid>
	userId, err := uuid.Parse(userIdStr)

	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetUserById(r.Context(), userId)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			http.Error(w, "User not found", http.StatusNotFound)
			return
		default:
			log.Printf("handler: error fetching user by ID: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Printf("handler: error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}

}
