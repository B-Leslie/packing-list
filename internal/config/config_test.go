package config

import (
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	getenv := func(string) string { return "" }
	_, err := Load(getenv)
	if err == nil {
		t.Fatal("expected error when BASE_URL is empty, got nil")
	}
}

func TestLoadHappy(t *testing.T) {
	env := map[string]string{
		"BASE_URL": "http://localhost:8080",
	}
	cfg, err := Load(func(k string) string { return env[k] })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("Port: got %q, want %q", cfg.Port, "8080")
	}
	if cfg.DataDir != "/data" {
		t.Errorf("DataDir: got %q, want %q", cfg.DataDir, "/data")
	}
	if cfg.BaseURL != "http://localhost:8080" {
		t.Errorf("BaseURL: got %q, want %q", cfg.BaseURL, "http://localhost:8080")
	}
	if cfg.SMTP.Configured() {
		t.Error("SMTP should not be configured without SMTP_HOST")
	}
}

func TestLoadSMTP(t *testing.T) {
	env := map[string]string{
		"BASE_URL":  "http://h",
		"SMTP_HOST": "smtp.example.com",
		"SMTP_USER": "u",
		"SMTP_PASS": "p",
	}
	cfg, err := Load(func(k string) string { return env[k] })
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !cfg.SMTP.Configured() {
		t.Fatal("expected SMTP configured")
	}
	if cfg.SMTP.From == "" {
		t.Error("expected SMTP.From to default to no-reply@<host>")
	}
}
