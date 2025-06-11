package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthStore struct {
	db *pgxpool.Pool
}

func (s *AuthStore) GetHashedPasswordAndId(username string) (userId, hashedPassw string, err error) {
	query := `SELECT password_hash FROM users WHERE username = $1`
	row := s.db.QueryRow(context.Background(), query, username, hashedPassw)

	if err := row.Scan(&userId, &hashedPassw); err != nil {
		if err == pgx.ErrNoRows {
			return "", "", errors.New("invalid credentials")
		}
		return "", "", err
	}

	return userId, hashedPassw, nil
}
