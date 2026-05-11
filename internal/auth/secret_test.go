package auth

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestLoadOrCreateSecretPersists(t *testing.T) {
	dir := t.TempDir()
	a, err := LoadOrCreateSecret(dir, "")
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	if len(a) != 32 {
		t.Errorf("expected 32 bytes, got %d", len(a))
	}
	b, err := LoadOrCreateSecret(dir, "")
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if !bytes.Equal(a, b) {
		t.Error("expected secret to be stable across calls")
	}
}

func TestLoadOrCreateSecretEnvOverrides(t *testing.T) {
	dir := t.TempDir()
	// Hex of 32 bytes = 64 chars.
	envHex := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	got, err := LoadOrCreateSecret(dir, envHex)
	if err != nil {
		t.Fatalf("env: %v", err)
	}
	if len(got) != 32 {
		t.Errorf("env path: expected 32 bytes, got %d", len(got))
	}
	// File must not have been created when env supplies it.
	_, err = readFileBytes(filepath.Join(dir, ".session_secret"))
	if err == nil {
		t.Error("file should not exist when SESSION_SECRET is provided")
	}
}
