package email

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/aeroxe/approval-flow/internal/config"
)

// Sender sends transactional emails over SMTP using the standard library.
// It is a thin, dependency-free wrapper around net/smtp with STARTTLS.
type Sender struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Sender {
	return &Sender{cfg: cfg}
}

// Configured reports whether SMTP credentials are available.
func (s *Sender) Configured() bool {
	return s.cfg.SMTPHost != ""
}

// Send delivers a plain-text (plus optional HTML) email to a single recipient.
// It returns an error only when SMTP is configured but delivery fails; callers
// decide whether a failure is fatal.
func (s *Sender) Send(to, subject, textBody string) error {
	if !s.Configured() {
		return fmt.Errorf("SMTP not configured")
	}
	if strings.TrimSpace(to) == "" {
		return fmt.Errorf("empty recipient")
	}

	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)

	var auth smtp.Auth
	if s.cfg.SMTPUser != "" {
		auth = smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, s.cfg.SMTPHost)
	}

	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n",
		s.cfg.SMTPFrom, to, subject, textBody,
	))

	if err := smtp.SendMail(addr, auth, s.cfg.SMTPFrom, []string{to}, msg); err != nil {
		return fmt.Errorf("smtp send failed: %w", err)
	}
	return nil
}
