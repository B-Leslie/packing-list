package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSessionsIssueLookup(t *testing.T) {
	db := newTestDB(t)
	if _, err := db.Exec(`INSERT INTO users(id,email) VALUES('u_a','a@example.com')`); err != nil {
		t.Fatal(err)
	}
	s := NewSessions(db, []byte("test-secret-32-bytes-padding-xx"), func() time.Time { return time.Now() })
	ctx := context.Background()

	cookieVal, err := s.Issue(ctx, "u_a")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	uid, err := s.Lookup(ctx, cookieVal)
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if uid != "u_a" {
		t.Errorf("uid: got %q", uid)
	}
}

func TestSessionsRejectTamperedCookie(t *testing.T) {
	db := newTestDB(t)
	if _, err := db.Exec(`INSERT INTO users(id,email) VALUES('u_a','a@example.com')`); err != nil {
		t.Fatal(err)
	}
	s := NewSessions(db, []byte("test-secret-32-bytes-padding-xx"), func() time.Time { return time.Now() })
	c, _ := s.Issue(context.Background(), "u_a")
	// Flip a byte.
	tampered := c[:len(c)-1] + "x"
	if _, err := s.Lookup(context.Background(), tampered); !errors.Is(err, ErrInvalidSession) {
		t.Errorf("expected ErrInvalidSession on tamper, got %v", err)
	}
}

func TestSessionsRevoke(t *testing.T) {
	db := newTestDB(t)
	db.Exec(`INSERT INTO users(id,email) VALUES('u_a','a@example.com')`)
	s := NewSessions(db, []byte("test-secret-32-bytes-padding-xx"), func() time.Time { return time.Now() })
	c, _ := s.Issue(context.Background(), "u_a")
	if err := s.Revoke(context.Background(), c); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if _, err := s.Lookup(context.Background(), c); !errors.Is(err, ErrInvalidSession) {
		t.Error("expected lookup to fail after revoke")
	}
}
