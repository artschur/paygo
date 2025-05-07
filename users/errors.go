package users

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrNoUsersFound = errors.New("no users found")
)
