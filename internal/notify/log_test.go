package notify

import (
	"bytes"
	"strings"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func TestLogNotifySkipsEmptyDiff(t *testing.T) {
	var buf bytes.Buffer
	n := NewLogNotifier(&buf, "")
	if err := n.Notify(ports.Diff{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output for empty diff, got %q", buf.String())
	}
}

func TestLogNotifyOpened(t *testing.T) {
	var buf bytes.Buffer
	n := NewLogNotifier(&buf, "test")
	err := n.Notify(ports.Diff{Opened: []int{8080}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "OPENED") || !strings.Contains(out, "8080") {
		t.Errorf("expected OPENED 8080 in output, got %q", out)
	}
	if !strings.Contains(out, "[test]") {
		t.Errorf("expected prefix [test] in output, got %q", out)
	}
}

func TestLogNotifyClosed(t *testing.T) {
	var buf bytes.Buffer
	n := NewLogNotifier(&buf, "pw")
	err := n.Notify(ports.Diff{Closed: []int{22}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "CLOSED") || !strings.Contains(out, "22") {
		t.Errorf("expected CLOSED 22 in output, got %q", out)
	}
}

func TestLogNotifyDefaultsNilWriter(t *testing.T) {
	// Should not panic when w is nil (falls back to stdout).
	n := NewLogNotifier(nil, "")
	if n.out == nil {
		t.Error("expected non-nil writer")
	}
	if n.prefix == "" {
		t.Error("expected non-empty default prefix")
	}
}

func TestLogNotifyBothDirections(t *testing.T) {
	var buf bytes.Buffer
	n := NewLogNotifier(&buf, "x")
	err := n.Notify(ports.Diff{Opened: []int{443}, Closed: []int{80}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "443") || !strings.Contains(out, "80") {
		t.Errorf("expected both ports in output, got %q", out)
	}
}
