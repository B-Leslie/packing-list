package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/bejl/packing-list/internal/ids"
)

var ErrInvalidSession = errors.New("invalid session")

const SessionTTL = 30 * 24 * time.Hour

type Sessions struct {
	db     *sql.DB
	secret []byte
	now    func() time.Time
}

func NewSessions(db *sql.DB, secret []byte, now func() time.Time) *Sessions {
	return &Sessions{db: db, secret: secret, now: now}
}

// Issue creates a session row and returns the signed cookie value: "<sid>.<hmac>".
func (s *Sessions) Issue(ctx context.Context, userID string) (string, error) {
	sid := ids.New()
	expires := s.now().Add(SessionTTL)
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO sessions(id,user_id,expires_at) VALUES (?,?,?)`,
		sid, userID, expires); err != nil {
		return "", err
	}
	return s.sign(sid), nil
}

// Lookup parses the cookie, validates HMAC, returns user ID. Side-effect:
// extends expires_at if more than half the TTL has elapsed (sliding renewal).
func (s *Sessions) Lookup(ctx context.Context, cookieVal string) (string, error) {
	sid, ok := s.verify(cookieVal)
	if !ok {
		return "", ErrInvalidSession
	}
	var userID string
	var expires time.Time
	err := s.db.QueryRowContext(ctx,
		`SELECT user_id, expires_at FROM sessions WHERE id = ?`, sid).
		Scan(&userID, &expires)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrInvalidSession
	}
	if err != nil {
		return "", err
	}
	now := s.now()
	if now.After(expires) {
		return "", ErrInvalidSession
	}
	if expires.Sub(now) < SessionTTL/2 {
		_, _ = s.db.ExecContext(ctx,
			`UPDATE sessions SET expires_at = ? WHERE id = ?`,
			now.Add(SessionTTL), sid)
	}
	return userID, nil
}

func (s *Sessions) Revoke(ctx context.Context, cookieVal string) error {
	sid, ok := s.verify(cookieVal)
	if !ok {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE id = ?`, sid)
	return err
}

func (s *Sessions) PurgeExpired(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at < ?`, s.now())
	return err
}

func (s *Sessions) sign(sid string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(sid))
	return sid + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *Sessions) verify(cookieVal string) (string, bool) {
	dot := strings.LastIndexByte(cookieVal, '.')
	if dot < 0 {
		return "", false
	}
	sid := cookieVal[:dot]
	got, err := base64.RawURLEncoding.DecodeString(cookieVal[dot+1:])
	if err != nil {
		return "", false
	}
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(sid))
	want := mac.Sum(nil)
	if !hmac.Equal(got, want) {
		return "", false
	}
	return sid, true
}
