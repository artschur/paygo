package users

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"paygo/models"

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

func (h *UserHandler) CreateUser(w http.ResponseWriter, r http.Request) {
	var newUser models.CreateUser
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		log.Printf("failed decoding json: %v", err)
		http.Error(w, "Internal server Error", http.StatusInternalServerError)
		return
	}
	if newUser.Email == "" || newUser.Name == "" || newUser.Password == "" {
		http.Error(w, "email, name and password are required", http.StatusBadRequest)
		return
	}
	createdUser, err := h.userService.CreateUser(r.Context(), newUser)
	if err != nil {
		log.Printf("Failed creating user: %v", err)
		http.Error(w, "failed writing in response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(createdUser)
	if err != nil {
		log.Printf("error encoding user response: %v", err)
		http.Error(w, "error encoding user response", http.StatusInternalServerError)
		return
	}
}
