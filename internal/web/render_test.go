package web

import (
	"bytes"
	"strings"
	"testing"
)

func TestRendererRendersLogin(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	var buf bytes.Buffer
	if err := r.Render(&buf, "login", map[string]any{"Title": "Sign in"}); err != nil {
		t.Fatalf("render: %v", err)
	}
	s := buf.String()
	if !strings.Contains(s, "Send magic link") {
		t.Errorf("expected login button text, got: %s", s)
	}
	if !strings.Contains(s, "<title>Sign in") {
		t.Errorf("expected title, got: %s", s)
	}
}
