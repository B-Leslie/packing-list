package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/bejl/packing-list/internal/ids"
)

var ErrInvalidToken = errors.New("invalid or expired token")

type Magic struct {
	db      *sql.DB
	mailer  Mailer
	baseURL string
	now     func() time.Time
}

func NewMagic(db *sql.DB, mailer Mailer, baseURL string, now func() time.Time) *Magic {
	return &Magic{db: db, mailer: mailer, baseURL: baseURL, now: now}
}

// Issue creates a single-use 15-minute token and sends a link to email.
func (m *Magic) Issue(ctx context.Context, email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return errors.New("email required")
	}
	// 32 random bytes -> base64url -> 43 chars.
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return err
	}
	token := base64.RawURLEncoding.EncodeToString(raw[:])
	hash := sha256.Sum256([]byte(token))
	id := ids.New()
	expires := m.now().Add(15 * time.Minute)
	if _, err := m.db.ExecContext(ctx,
		`INSERT INTO magic_tokens(id,email,token_hash,expires_at) VALUES (?,?,?,?)`,
		id, email, hash[:], expires); err != nil {
		return err
	}
	link := m.baseURL + "/auth/verify?t=" + token
	return m.mailer.SendMagicLink(ctx, email, link)
}

// Consume validates a token and marks it used. Returns the associated email.
func (m *Magic) Consume(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", ErrInvalidToken
	}
	hash := sha256.Sum256([]byte(token))
	var (
		id, email string
		expires   time.Time
		used      sql.NullTime
	)
	err := m.db.QueryRowContext(ctx,
		`SELECT id,email,expires_at,used_at FROM magic_tokens WHERE token_hash = ?`, hash[:]).
		Scan(&id, &email, &expires, &used)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrInvalidToken
	}
	if err != nil {
		return "", err
	}
	if used.Valid {
		return "", ErrInvalidToken
	}
	if m.now().After(expires) {
		return "", ErrInvalidToken
	}
	res, err := m.db.ExecContext(ctx,
		`UPDATE magic_tokens SET used_at = CURRENT_TIMESTAMP WHERE id = ? AND used_at IS NULL`, id)
	if err != nil {
		return "", err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return "", ErrInvalidToken
	}
	return email, nil
}

// PurgeExpired hard-deletes tokens older than now-1h. Safe to call periodically.
func (m *Magic) PurgeExpired(ctx context.Context) error {
	_, err := m.db.ExecContext(ctx,
		`DELETE FROM magic_tokens WHERE expires_at < ?`, m.now().Add(-time.Hour))
	return err
}
