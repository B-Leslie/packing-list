package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"

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
	if m.Cfg.Insecure {
		return sendPlain(addr, m.Cfg.From, email, []byte(body))
	}
	var auth smtp.Auth
	if m.Cfg.User != "" {
		auth = smtp.PlainAuth("", m.Cfg.User, m.Cfg.Pass, m.Cfg.Host)
	}
	return smtp.SendMail(addr, auth, m.Cfg.From, []string{email}, []byte(body))
}

// sendPlain mirrors smtp.SendMail but skips StartTLS, for LAN relays
// whose self-signed certs would fail verification. Auth is intentionally
// omitted -- never send credentials over an unencrypted channel.
func sendPlain(addr, from, to string, msg []byte) error {
	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer c.Close()
	host := addr
	if i := strings.LastIndex(host, ":"); i >= 0 {
		host = host[:i]
	}
	if err := c.Hello(host); err != nil {
		return err
	}
	if err := c.Mail(from); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return c.Quit()
}

// NewMailer returns SMTPMailer if SMTP is configured, otherwise LogMailer.
func NewMailer(cfg config.Config, logger *slog.Logger) Mailer {
	if cfg.SMTP.Configured() {
		return &SMTPMailer{Cfg: cfg.SMTP}
	}
	return &LogMailer{Logger: logger}
}
