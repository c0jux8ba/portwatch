package daemon

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/user/portwatch/internal/config"
	"github.com/user/portwatch/internal/ports"
)

// mockNotifier records diffs it receives.
type mockNotifier struct {
	calls []ports.Diff
}

func (m *mockNotifier) Notify(d ports.Diff) error {
	m.calls = append(m.calls, d)
	return nil
}

func defaultTestConfig(start, end int) *config.Config {
	cfg := config.DefaultConfig()
	cfg.StartPort = start
	cfg.EndPort = end
	cfg.IntervalSecs = 1
	cfg.ScanTimeoutMs = 100
	return cfg
}

func TestDaemonBaselineDoesNotAlert(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer ln.Close()

	port := ln.Addr().(*net.TCPAddr).Port
	cfg := defaultTestConfig(port, port)
	notifier := &mockNotifier{}
	d := New(cfg, notifier)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = d.Run(ctx)

	// No tick should have fired in 200ms with a 1s interval, so no alerts.
	if len(notifier.calls) != 0 {
		t.Errorf("expected 0 notifications during baseline, got %d", len(notifier.calls))
	}
}

func TestDaemonDetectsNewPort(t *testing.T) {
	cfg := defaultTestConfig(19900, 19910)
	notifier := &mockNotifier{}
	d := New(cfg, notifier)

	// Establish empty baseline manually.
	d.previous = []int{}

	// Open a port inside the range.
	ln, err := net.Listen("tcp", "127.0.0.1:19905")
	if err != nil {
		t.Skipf("port 19905 unavailable: %v", err)
	}
	defer ln.Close()

	if err := d.tick(); err != nil {
		t.Fatalf("tick error: %v", err)
	}

	if len(notifier.calls) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notifier.calls))
	}
	if len(notifier.calls[0].Opened) == 0 {
		t.Errorf("expected opened ports in diff, got none")
	}
}
