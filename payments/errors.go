package payments

import "errors"

var (
	ErrNoPaymentsFound = errors.New("No payments found in database")
	ErrUserIdNotFound  = errors.New("No user found with the ID passed")
)
