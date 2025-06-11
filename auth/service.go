package auth

import (
	"errors"
	"fmt"
	"paygo/utils"
)

type AuthStoreInterface interface {
	GetHashedPassword(username string) (userId, hashedPassw string, err error)
	Register(username, email, password string) (string, error)
}

type AuthService struct {
	store AuthStoreInterface
}

func NewAuthService(store AuthStoreInterface) *AuthService {
	return &AuthService{
		store: store,
	}
}

func (s *AuthService) Login(username, password string) (string, error) {

	userId, hashedPass, err := s.store.GetHashedPassword(username)
	if err != nil {
		return "", err
	}

	ok, err := utils.CheckPassword(hashedPass, password)
	if err != nil {
		return "", fmt.Errorf("error checking password: %v", err)
	}

	if !ok {
		return "", errors.New("Invalid password")
	}

	return userId, nil
}

func (s *AuthService) Register(username, email, password string) (string, error) {
	// s.store.
	// create a user return id and create a wallet from it
	return "", errors.New("Register method not implemented")
}
