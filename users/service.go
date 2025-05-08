package users

import (
	"context"
	"errors"
	"fmt"
	"paygo/models"

	"github.com/google/uuid"
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

func (s *UserService) GetUserById(ctx context.Context, userId uuid.UUID) (user models.User, err error) {
	user, err = s.userStore.GetUserById(ctx, userId)

	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return models.User{}, err
		}
		return models.User{}, fmt.Errorf("service: failed to get user by id: %v", err.Error())
	}

	return user, nil

}
