package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

type captureMailer struct {
	gotEmail, gotLink string
}

func (c *captureMailer) SendMagicLink(_ context.Context, email, link string) error {
	c.gotEmail = email
	c.gotLink = link
	return nil
}

func TestIssueAndConsumeRoundtrip(t *testing.T) {
	db := newTestDB(t)
	cap := &captureMailer{}
	mt := NewMagic(db, cap, "https://app", func() time.Time { return time.Now() })
	ctx := context.Background()

	if err := mt.Issue(ctx, "alice@example.com"); err != nil {
		t.Fatalf("issue: %v", err)
	}
	if cap.gotEmail != "alice@example.com" {
		t.Errorf("captured email: %s", cap.gotEmail)
	}
	// Extract token from link.
	const prefix = "https://app/auth/verify?t="
	if len(cap.gotLink) <= len(prefix) {
		t.Fatalf("bad link: %s", cap.gotLink)
	}
	token := cap.gotLink[len(prefix):]
	email, err := mt.Consume(ctx, token)
	if err != nil {
		t.Fatalf("consume: %v", err)
	}
	if email != "alice@example.com" {
		t.Errorf("consume email: %s", email)
	}
}

func TestConsumeRejectsReuse(t *testing.T) {
	db := newTestDB(t)
	cap := &captureMailer{}
	mt := NewMagic(db, cap, "https://app", func() time.Time { return time.Now() })
	ctx := context.Background()
	_ = mt.Issue(ctx, "x@example.com")
	token := cap.gotLink[len("https://app/auth/verify?t="):]
	if _, err := mt.Consume(ctx, token); err != nil {
		t.Fatalf("first consume: %v", err)
	}
	if _, err := mt.Consume(ctx, token); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken on reuse, got %v", err)
	}
}

func TestConsumeRejectsExpired(t *testing.T) {
	db := newTestDB(t)
	cap := &captureMailer{}
	// Fixed clock at t0; expired tokens are >15 min old.
	t0 := time.Now()
	mt := NewMagic(db, cap, "https://app", func() time.Time { return t0 })
	ctx := context.Background()
	_ = mt.Issue(ctx, "x@example.com")
	token := cap.gotLink[len("https://app/auth/verify?t="):]

	mt.now = func() time.Time { return t0.Add(20 * time.Minute) }
	if _, err := mt.Consume(ctx, token); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken after expiry, got %v", err)
	}
}
