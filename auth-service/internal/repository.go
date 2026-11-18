package internal

import (
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (repo *Repository) CreateUser(ctx context.Context, username, hash string) (User, error) {
	var user User
	err := repo.db.QueryRowContext(ctx,
		`INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id, username, created_at`,
		username, hash,
	).Scan(&user.ID, &user.Username, &user.CreatedAt)
	return user, err
}

func (repo *Repository) Credentials(ctx context.Context, username string) (User, string, error) {
	var user User
	var passwordHash string
	err := repo.db.QueryRowContext(ctx,
		`SELECT id, username, password_hash, created_at FROM users WHERE username = $1`,
		username,
	).Scan(&user.ID, &user.Username, &passwordHash, &user.CreatedAt)
	return user, passwordHash, err
}

func (repo *Repository) FindUser(ctx context.Context, username string) (User, error) {
	var user User
	err := repo.db.QueryRowContext(ctx,
		`SELECT id, username, created_at FROM users WHERE username = $1`,
		username,
	).Scan(&user.ID, &user.Username, &user.CreatedAt)
	return user, err
}
