package daemon

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/user/portwatch/internal/config"
	"github.com/user/portwatch/internal/notify"
	"github.com/user/portwatch/internal/ports"
)

func listenTCP(t *testing.T) (int, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	return port, func() { ln.Close() }
}

func makeRunner(t *testing.T, from, to int) (*Runner, *capNotifier) {
	t.Helper()
	cfg := &config.Config{PortRange: config.PortRange{From: from, To: to}}
	scanner := ports.NewScanner(cfg)
	baseline, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("baseline scan: %v", err)
	}
	snap := ports.NewSnapshot(baseline)
	cap := &capNotifier{}
	return NewRunner(scanner, cap, snap), cap
}

type capNotifier struct {
	calls []ports.Diff
}

func (c *capNotifier) Notify(d ports.Diff) error {
	c.calls = append(c.calls, d)
	return nil
}

func TestRunnerTickNoChange(t *testing.T) {
	port, close := listenTCP(t)
	defer close()
	runner, cap := makeRunner(t, port, port)
	changed, err := runner.Tick(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if changed {
		t.Error("expected no change on second scan of same state")
	}
	if len(cap.calls) != 0 {
		t.Errorf("expected 0 notifications, got %d", len(cap.calls))
	}
}

func TestRunnerTickDetectsNewPort(t *testing.T) {
	port, close := listenTCP(t)
	runner, cap := makeRunner(t, port, port)
	close() // close before tick so port disappears

	changed, err := runner.Tick(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Error("expected change to be detected")
	}
	if len(cap.calls) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(cap.calls))
	}
	if len(cap.calls[0].Closed) != 1 || cap.calls[0].Closed[0] != port {
		t.Errorf("expected closed port %d, got %v", port, cap.calls[0].Closed)
	}
}

func TestRunLoopCancelStops(t *testing.T) {
	port, close := listenTCP(t)
	defer close()
	runner, _ := makeRunner(t, port, port)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	err := runner.RunLoop(ctx, 10*time.Millisecond)
	if err != nil {
		t.Errorf("RunLoop returned error: %v", err)
	}
}
