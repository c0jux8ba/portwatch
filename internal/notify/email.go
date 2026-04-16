package notify

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/user/portwatch/internal/ports"
)

// EmailConfig holds SMTP configuration.
type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	To       []string
}

type emailNotifier struct {
	cfg       EmailConfig
	formatter *Formatter
	sendMail  func(addr string, a smtp.Auth, from string, to []string, msg []byte) error
}

// NewEmailNotifier creates a Notifier that sends alerts via SMTP.
func NewEmailNotifier(cfg EmailConfig) Notifier {
	return &emailNotifier{
		cfg:       cfg,
		formatter: NewFormatter(),
		sendMail:  smtp.SendMail,
	}
}

func (e *emailNotifier) Notify(d ports.Diff) error {
	if d.IsEmpty() {
		return nil
	}

	subject := "portwatch: port change detected"
	body := e.formatter.Format(d)

	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		e.cfg.From,
		strings.Join(e.cfg.To, ", "),
		subject,
		body,
	))

	addr := fmt.Sprintf("%s:%d", e.cfg.Host, e.cfg.Port)
	var auth smtp.Auth
	if e.cfg.Username != "" {
		auth = smtp.PlainAuth("", e.cfg.Username, e.cfg.Password, e.cfg.Host)
	}

	return e.sendMail(addr, auth, e.cfg.From, e.cfg.To, msg)
}
