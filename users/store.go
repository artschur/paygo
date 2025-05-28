package users

import (
	"context"
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
	wantCols := []string{"u.id", "u.name", "u.email", "u.created_at", "w.id", "w.balance", "w.currency", "w.updated_at"}
	query := fmt.Sprintf(
		`
		SELECT %s
	 	FROM users u
		LEFT JOIN wallets w ON u.id = w.user_id
		WHERE u.id = $1`,
		strings.Join(wantCols, ", "),
	)
	err = s.db.QueryRow(ctx, query, userId).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
		&user.Wallet.ID,
		&user.Wallet.Balance,
		&user.Wallet.Currency,
		&user.Wallet.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return user, fmt.Errorf("store: user not found: %w", ErrUserNotFound)
		}
		return user, fmt.Errorf("store: failed to fetch user by ID: %w", err)
	}
	return user, nil
}

func (s *UserStore) GetAllUsers(ctx context.Context) (users []models.User, err error) {
	wantCols := []string{"id", "name", "email", "created_at"}
	query := fmt.Sprintf(
		"SELECT %s FROM users ORDER BY created_at DESC",
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

	if rows.Err() != nil {
		return nil, fmt.Errorf("store: error iterating over rows: %w", rows.Err())
	}
	return usersContainer, nil
}

func (s *UserStore) CreateUser(ctx context.Context, newUser *models.CreateUser) (createdUser models.User, err error) {
	row := s.db.QueryRow(ctx, `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, email, name, created_at;
		`, newUser.Name, newUser.Email, newUser.Password)

	err = row.Scan(&createdUser.ID, &createdUser.Email, &createdUser.Name, &createdUser.CreatedAt)
	if err != nil {
		return models.User{}, fmt.Errorf("store: failed to insert user %v ", err)
	}

	return createdUser, nil
}
