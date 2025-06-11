package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type AuthServiceInterface interface {
	Login(username, password string) (string, error)
	Register(username, email, password string) (string, error)
}

type AuthHandler struct {
	service AuthServiceInterface
}

func NewAuthHandler(s AuthServiceInterface) *AuthHandler {
	return &AuthHandler{
		service: s,
	}
}

func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if creds.Username == "" || creds.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	user_id, err := h.service.Login(creds.Username, creds.Password)
	if err != nil {
		http.Error(w, fmt.Sprintf("Login failed: %v", err), http.StatusUnauthorized)
		return
	}

	token, err := CreateToken(creds.Username, user_id)
	r.Header.Add("Authorization", token)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if creds.Username == "" || creds.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}
	user_id, err := h.service.Register(creds.Username, creds.Email, creds.Password)
	if err != nil {
		http.Error(w, fmt.Sprintf("Registration failed: %v", err), http.StatusInternalServerError)
		return
	}
	token, err := CreateToken(creds.Username, user_id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating token: %v", err), http.StatusInternalServerError)
		return
	}
	r.Header.Add("Authorization", token)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
