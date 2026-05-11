package auth

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestLogMailerWritesLink(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	m := &LogMailer{Logger: logger}
	if err := m.SendMagicLink(context.Background(), "alice@example.com", "https://app/auth/verify?t=abc"); err != nil {
		t.Fatalf("send: %v", err)
	}
	if !strings.Contains(buf.String(), "alice@example.com") {
		t.Errorf("log missing email: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "https://app/auth/verify?t=abc") {
		t.Errorf("log missing link: %q", buf.String())
	}
}
