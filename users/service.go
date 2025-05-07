package users

import (
	"context"
	"fmt"
	"paygo/models"
)

type UserService struct {
	userStore *UserStore
}

func NewUserService(userStore *UserStore) *UserService {
	return &UserService{
		userStore,
	}
}

func (s *UserService) GetAllUsers(ctx context.Context) (users []models.User, err error) {
	users, err = s.userStore.GetAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get all users: %w:", err)
	}
	if len(users) == 0 {
		return nil, ErrNoUsersFound
	}

	return users, nil
}
