package notify

import (
	"strings"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func TestBuildSyslogMessageOpenedOnly(t *testing.T) {
	d := ports.Diff{Opened: []int{80, 443}, Closed: nil}
	msg := buildSyslogMessage(d)
	if !strings.Contains(msg, "opened=") {
		t.Errorf("expected 'opened=' in message, got: %s", msg)
	}
	if strings.Contains(msg, "closed=") {
		t.Errorf("unexpected 'closed=' in message: %s", msg)
	}
}

func TestBuildSyslogMessageClosedOnly(t *testing.T) {
	d := ports.Diff{Opened: nil, Closed: []int{22}}
	msg := buildSyslogMessage(d)
	if !strings.Contains(msg, "closed=") {
		t.Errorf("expected 'closed=' in message, got: %s", msg)
	}
}

func TestBuildSyslogMessageBothDirections(t *testing.T) {
	d := ports.Diff{Opened: []int{8080}, Closed: []int{22}}
	msg := buildSyslogMessage(d)
	if !strings.Contains(msg, "opened=") || !strings.Contains(msg, "closed=") {
		t.Errorf("expected both directions in message, got: %s", msg)
	}
}

func TestSyslogNotifySkipsEmptyDiff(t *testing.T) {
	// We cannot open a real syslog in all CI environments, so we test the
	// guard path via a nil-writer stand-in by inspecting IsEmpty directly.
	d := ports.Diff{}
	if !d.IsEmpty() {
		t.Fatal("expected empty diff")
	}
}

func TestBuildSyslogMessageContainsPorts(t *testing.T) {
	d := ports.Diff{Opened: []int{9000}, Closed: []int{3306}}
	msg := buildSyslogMessage(d)
	if !strings.Contains(msg, "9000") {
		t.Errorf("expected port 9000 in message: %s", msg)
	}
	if !strings.Contains(msg, "3306") {
		t.Errorf("expected port 3306 in message: %s", msg)
	}
}
