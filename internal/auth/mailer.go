package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"

	"github.com/bejl/packing-list/internal/config"
)

// Mailer abstracts magic-link delivery so we can swap log + smtp + tests.
type Mailer interface {
	SendMagicLink(ctx context.Context, email, link string) error
}

// LogMailer prints the link to slog at INFO. Use in dev / when SMTP unset.
type LogMailer struct {
	Logger *slog.Logger
}

func (m *LogMailer) SendMagicLink(_ context.Context, email, link string) error {
	m.Logger.Info("magic link issued (no SMTP configured)",
		"email", email, "link", link)
	return nil
}

// SMTPMailer sends a minimal RFC 5322 message via PLAIN SMTP AUTH.
type SMTPMailer struct {
	Cfg config.SMTP
}

func (m *SMTPMailer) SendMagicLink(_ context.Context, email, link string) error {
	addr := m.Cfg.Host + ":" + m.Cfg.Port
	body := fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: Sign in to Packing List\r\n\r\nOpen this link to sign in. It expires in 15 minutes.\r\n\r\n%s\r\n",
		email, m.Cfg.From, link)
	var auth smtp.Auth
	if m.Cfg.User != "" {
		auth = smtp.PlainAuth("", m.Cfg.User, m.Cfg.Pass, m.Cfg.Host)
	}
	return smtp.SendMail(addr, auth, m.Cfg.From, []string{email}, []byte(body))
}

// NewMailer returns SMTPMailer if SMTP is configured, otherwise LogMailer.
func NewMailer(cfg config.Config, logger *slog.Logger) Mailer {
	if cfg.SMTP.Configured() {
		return &SMTPMailer{Cfg: cfg.SMTP}
	}
	return &LogMailer{Logger: logger}
}
