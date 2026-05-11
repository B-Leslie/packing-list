// Package auth covers users, magic-link tokens, sessions, CSRF.
package auth

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/bejl/packing-list/internal/ids"
)

var ErrNotFound = errors.New("not found")

type User struct {
	ID    string
	Email string
}

type Users struct{ db *sql.DB }

func NewUsers(db *sql.DB) *Users { return &Users{db: db} }

// FindOrCreate returns (id, created, err). Match is case-insensitive on email.
func (u *Users) FindOrCreate(ctx context.Context, email string) (string, bool, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return "", false, errors.New("email required")
	}
	// Try existing first.
	var id string
	err := u.db.QueryRowContext(ctx, `SELECT id FROM users WHERE email = ? AND deleted_at IS NULL`, email).Scan(&id)
	if err == nil {
		return id, false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", false, err
	}
	id = ids.New()
	_, err = u.db.ExecContext(ctx, `INSERT INTO users(id,email) VALUES (?,?)`, id, email)
	if err != nil {
		return "", false, err
	}
	return id, true, nil
}

func (u *Users) Get(ctx context.Context, id string) (User, error) {
	var usr User
	err := u.db.QueryRowContext(ctx, `SELECT id,email FROM users WHERE id = ? AND deleted_at IS NULL`, id).Scan(&usr.ID, &usr.Email)
	if errors.Is(err, sql.ErrNoRows) {
		return usr, ErrNotFound
	}
	return usr, err
}
