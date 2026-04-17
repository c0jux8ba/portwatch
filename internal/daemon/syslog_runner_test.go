package daemon

import (
	"testing"

	"github.com/user/portwatch/internal/ports"
)

// TestSyslogRunnerSkipsEmptyDiff verifies that an empty diff does not cause an
// error when forwarded to the syslog notifier.
func TestSyslogRunnerSkipsEmptyDiff(t *testing.T) {
	// NewSyslogNotifier requires a real syslog socket; skip if unavailable.
	cfg := defaultTestConfig()
	r, err := makeRunner(cfg)
	if err != nil {
		t.Skipf("makeRunner: %v", err)
	}
	sr, err := NewSyslogRunner(r, "portwatch-test")
	if err != nil {
		t.Skipf("syslog unavailable: %v", err)
	}
	defer sr.Close()

	d := ports.Diff{}
	if err := sr.NotifySyslog(d); err != nil {
		t.Errorf("expected no error for empty diff, got: %v", err)
	}
}

// TestSyslogRunnerNotifiesChange verifies that a non-empty diff is forwarded
// without error when syslog is available.
func TestSyslogRunnerNotifiesChange(t *testing.T) {
	cfg := defaultTestConfig()
	r, err := makeRunner(cfg)
	if err != nil {
		t.Skipf("makeRunner: %v", err)
	}
	sr, err := NewSyslogRunner(r, "portwatch-test")
	if err != nil {
		t.Skipf("syslog unavailable: %v", err)
	}
	defer sr.Close()

	d := ports.Diff{Opened: []int{9999}, Closed: []int{8888}}
	if err := sr.NotifySyslog(d); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
