package users

import (
	"context"
	"errors"
	"fmt"
	"paygo/models"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserStore struct {
	db *pgxpool.Pool
}

func NewUserStore(db *pgxpool.Pool) *UserStore {
	return &UserStore{
		db,
	}
}

func (s *UserStore) GetUserById(ctx context.Context, userId uuid.UUID) (user models.User, err error) {
	wantCols := []string{"id", "name", "email", "created_at"}
	query := fmt.Sprintf(
		"SELECT %s FROM users WHERE id = $1",
		strings.Join(wantCols, ", "),
	)
	err = s.db.QueryRow(ctx, query, userId).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return user, fmt.Errorf("store: user not found: %w", err)
		}
		return user, fmt.Errorf("store: failed to fetch user by ID: %w", err)
	}
	return user, nil
}

func (s *UserStore) GetAllUsers(ctx context.Context) (users []models.User, err error) {
	wantCols := []string{"id", "name", "email", "created_at"}
	query := fmt.Sprintf(
		"SELECT %s FROM users",
		strings.Join(wantCols, ", "),
	)
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("store: failed to fetch all users: %w", err)
	}

	defer rows.Close()

	var usersContainer []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("store: failed to scan user row: %w", err)
		}
		usersContainer = append(usersContainer, user)
	}
	if len(users) == 0 {
		return nil, errors.New("No users found")
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("store: error iterating over rows: %w", rows.Err())
	}

	return users, nil
}
