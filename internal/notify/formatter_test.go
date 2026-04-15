package notify

import (
	"strings"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func TestFormatterFormatOpened(t *testing.T) {
	f := NewFormatter("testhost")
	diff := ports.Diff{
		Opened: []int{80, 443},
		Closed: []int{},
	}
	msg := f.Format(diff)
	if !strings.Contains(msg, "80") {
		t.Errorf("expected port 80 in message, got: %s", msg)
	}
	if !strings.Contains(msg, "443") {
		t.Errorf("expected port 443 in message, got: %s", msg)
	}
	if !strings.Contains(msg, "testhost") {
		t.Errorf("expected hostname in message, got: %s", msg)
	}
}

func TestFormatterFormatClosed(t *testing.T) {
	f := NewFormatter("myhost")
	diff := ports.Diff{
		Opened: []int{},
		Closed: []int{22, 8080},
	}
	msg := f.Format(diff)
	if !strings.Contains(msg, "22") {
		t.Errorf("expected port 22 in message, got: %s", msg)
	}
	if !strings.Contains(msg, "8080") {
		t.Errorf("expected port 8080 in message, got: %s", msg)
	}
}

func TestFormatterFormatBothDirections(t *testing.T) {
	f := NewFormatter("host")
	diff := ports.Diff{
		Opened: []int{9000},
		Closed: []int{3000},
	}
	msg := f.Format(diff)
	if !strings.Contains(msg, "opened") && !strings.Contains(msg, "Opened") {
		t.Errorf("expected 'opened' keyword in message, got: %s", msg)
	}
	if !strings.Contains(msg, "closed") && !strings.Contains(msg, "Closed") {
		t.Errorf("expected 'closed' keyword in message, got: %s", msg)
	}
}

func TestFormatterFormatEmptyDiff(t *testing.T) {
	f := NewFormatter("host")
	diff := ports.Diff{
		Opened: []int{},
		Closed: []int{},
	}
	msg := f.Format(diff)
	if msg != "" {
		t.Errorf("expected empty string for empty diff, got: %s", msg)
	}
}

func TestFormatterDefaultHostname(t *testing.T) {
	f := NewFormatter("")
	diff := ports.Diff{
		Opened: []int{80},
		Closed: []int{},
	}
	msg := f.Format(diff)
	if msg == "" {
		t.Error("expected non-empty message even with empty hostname")
	}
}
