package payments

import "errors"

var (
	ErrNoPaymentsFound      = errors.New("No payments found in database")
	ErrUserIdNotFound       = errors.New("No user found with the ID passed")
	ErrIllegalUserId        = errors.New("Illegal user ID provided")
	ErrDepositAmountInvalid = errors.New("Deposit amount must be greater than zero")
)
