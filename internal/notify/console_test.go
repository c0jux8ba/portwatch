package notify

import (
	"bytes"
	"strings"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func TestConsoleNotifySkipsEmptyDiff(t *testing.T) {
	var buf bytes.Buffer
	n := NewConsoleNotifier("", &buf)
	_ = n.Notify(ports.Diff{})
	if buf.Len() != 0 {
		t.Fatalf("expected no output for empty diff, got %q", buf.String())
	}
}

func TestConsoleNotifyOpened(t *testing.T) {
	var buf bytes.Buffer
	n := NewConsoleNotifier("pw", &buf)
	_ = n.Notify(ports.Diff{Opened: []int{8080}})
	out := buf.String()
	if !strings.Contains(out, "OPENED") {
		t.Errorf("expected OPENED in output, got %q", out)
	}
	if !strings.Contains(out, "8080") {
		t.Errorf("expected port 8080 in output, got %q", out)
	}
}

func TestConsoleNotifyClosed(t *testing.T) {
	var buf bytes.Buffer
	n := NewConsoleNotifier("pw", &buf)
	_ = n.Notify(ports.Diff{Closed: []int{443}})
	out := buf.String()
	if !strings.Contains(out, "CLOSED") {
		t.Errorf("expected CLOSED in output, got %q", out)
	}
}

func TestConsoleNotifyDefaultsNilWriter(t *testing.T) {
	n := NewConsoleNotifier("", nil)
	if n.out == nil {
		t.Fatal("expected non-nil writer")
	}
}

func TestConsoleNotifyDefaultPrefix(t *testing.T) {
	n := NewConsoleNotifier("", nil)
	if n.prefix != "portwatch" {
		t.Errorf("expected default prefix 'portwatch', got %q", n.prefix)
	}
}

func TestConsoleNotifyBothDirections(t *testing.T) {
	var buf bytes.Buffer
	n := NewConsoleNotifier("pw", &buf)
	_ = n.Notify(ports.Diff{Opened: []int{9000}, Closed: []int{22}})
	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d: %q", len(lines), out)
	}
}
