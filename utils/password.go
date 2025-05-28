package utils

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(rawPass string) (hashedPass string, err error) {
	passBytes, err := bcrypt.GenerateFromPassword([]byte(rawPass), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("failed to encrypt password")
	}
	return string(passBytes), nil
}

func CheckPassword(hashed_pass, raw_passw string) (is_ok bool, err error) {
	err = bcrypt.CompareHashAndPassword([]byte(hashed_pass), []byte(raw_passw))
	if err != nil {
		return false, err
	}

	return true, nil
}
