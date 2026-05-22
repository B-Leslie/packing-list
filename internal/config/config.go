// Package config parses runtime configuration from environment variables.
package config

import (
	"errors"
	"net/url"
	"strings"
)

type SMTP struct {
	Host string
	Port string
	User string
	Pass string
	From string
}

func (s SMTP) Configured() bool { return s.Host != "" }

type Config struct {
	Port            string
	BaseURL         string
	DataDir         string
	SessionSecret   string // empty means: load/generate at boot
	BootstrapEmail  string
	// AllowedEmails gates /login: only existing users can request a
	// magic link. The bootstrap step pre-creates a row for each entry
	// here so the first login attempt finds them. Trip-member invites
	// from existing users still create new rows independently.
	// Empty slice = open signup (legacy behaviour).
	AllowedEmails   []string
	SMTP            SMTP
}

// Load reads config using the given getenv function (so tests don't touch real env).
func Load(getenv func(string) string) (Config, error) {
	c := Config{
		Port:            or(getenv("PORT"), "8080"),
		BaseURL:         getenv("BASE_URL"),
		DataDir:         or(getenv("DATA_DIR"), "/data"),
		SessionSecret:   getenv("SESSION_SECRET"),
		BootstrapEmail:  getenv("BOOTSTRAP_EMAIL"),
		AllowedEmails:   parseEmailList(getenv("ALLOWED_EMAILS")),
		SMTP: SMTP{
			Host: getenv("SMTP_HOST"),
			Port: or(getenv("SMTP_PORT"), "587"),
			User: getenv("SMTP_USER"),
			Pass: getenv("SMTP_PASS"),
			From: getenv("SMTP_FROM"),
		},
	}
	if c.BaseURL == "" {
		return c, errors.New("BASE_URL is required")
	}
	u, err := url.Parse(c.BaseURL)
	if err != nil || u.Host == "" {
		return c, errors.New("BASE_URL is not a valid URL")
	}
	if c.SMTP.Configured() && c.SMTP.From == "" {
		c.SMTP.From = "no-reply@" + u.Hostname()
	}
	return c, nil
}

func or(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

// parseEmailList splits a comma-separated env value, trims whitespace,
// lowercases, and drops empties.
func parseEmailList(v string) []string {
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		e := strings.ToLower(strings.TrimSpace(p))
		if e != "" {
			out = append(out, e)
		}
	}
	return out
}
