package notify

import (
	"errors"
	"net/smtp"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func makeEmailNotifier(send func(string, smtp.Auth, string, []string, []byte) error) *emailNotifier {
	cfg := EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
		From:     "alerts@example.com",
		To:       []string{"admin@example.com"},
	}
	return &emailNotifier{cfg: cfg, formatter: NewFormatter(), sendMail: send}
}

func TestEmailNotifySkipsEmptyDiff(t *testing.T) {
	called := false
	n := makeEmailNotifier(func(_ string, _ smtp.Auth, _ string, _ []string, _ []byte) error {
		called = true
		return nil
	})
	_ = n.Notify(ports.Diff{})
	if called {
		t.Fatal("expected sendMail not to be called for empty diff")
	}
}

func TestEmailNotifySendsOnChange(t *testing.T) {
	var capturedMsg []byte
	n := makeEmailNotifier(func(_ string, _ smtp.Auth, _ string, _ []string, msg []byte) error {
		capturedMsg = msg
		return nil
	})
	d := ports.Diff{Opened: []int{8080}}
	if err := n.Notify(d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(capturedMsg) == 0 {
		t.Fatal("expected message to be sent")
	}
}

func TestEmailNotifyReturnsError(t *testing.T) {
	n := makeEmailNotifier(func(_ string, _ smtp.Auth, _ string, _ []string, _ []byte) error {
		return errors.New("smtp failure")
	})
	d := ports.Diff{Opened: []int{443}}
	if err := n.Notify(d); err == nil {
		t.Fatal("expected error from sendMail")
	}
}

func TestEmailNotifyNoAuthWhenUsernameEmpty(t *testing.T) {
	var capturedAuth smtp.Auth
	n := makeEmailNotifier(func(_ string, a smtp.Auth, _ string, _ []string, _ []byte) error {
		capturedAuth = a
		return nil
	})
	n.cfg.Username = ""
	d := ports.Diff{Closed: []int{22}}
	_ = n.Notify(d)
	if capturedAuth != nil {
		t.Fatal("expected nil auth when username is empty")
	}
}
