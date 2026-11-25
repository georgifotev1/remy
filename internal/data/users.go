package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// insertUser inserts u and fills in its ID and timestamps.
func insertUser(ctx context.Context, q DBTX, u *User) error {
	const query = `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, created_at, updated_at`

	err := q.QueryRow(ctx, query, u.Name, u.Email, u.PasswordHash).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if isUniqueViolation(err, "users_email_uq") {
		return ErrDuplicateEmail
	}
	return err
}

type UserModel struct {
	db *pgxpool.Pool
}

// GetByEmail returns ErrRecordNotFound if no user has that email.
// The email column is CITEXT, so the match is case-insensitive.
func (m *UserModel) GetByEmail(ctx context.Context, email string) (*User, error) {
	const query = `
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1`

	var u User
	err := m.db.QueryRow(ctx, query, email).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRecordNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetByID returns ErrRecordNotFound if no user has that id.
func (m *UserModel) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	const query = `
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1`

	var u User
	err := m.db.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRecordNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
