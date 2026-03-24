package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type UserRepository interface {
	GetUserByID(ctx context.Context, id string) (*User, error)
}

type postgresUserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &postgresUserRepo{pool: pool}
}

func (r *postgresUserRepo) GetUserByID(ctx context.Context, id string) (*User, error) {
	query := `SELECT id, email, name, created_at 
	FROM users 
	WHERE id = $1`

	var user User

	err := r.pool.QueryRow(ctx, query, id).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("database error fetching user %s: %w", id, err)
	}
	return &user, nil
}
