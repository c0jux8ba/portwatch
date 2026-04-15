package ports

import (
	"net"
	"testing"
	"time"
)

func openTCPListener(t *testing.T) (net.Listener, int) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to open listener: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	return ln, port
}

func TestWatcherEmitsEventOnNewPort(t *testing.T) {
	ln, port := openTCPListener(t)
	defer ln.Close()

	scanner := NewScanner(port, port, 50*time.Millisecond)
	baseline := NewBaseline("")
	// seed baseline with empty set so the watcher sees the port as new
	baseline.Set([]int{})

	w := NewWatcher(scanner, nil, baseline, 20*time.Millisecond)
	w.Start()
	defer w.Stop()

	select {
	case ev := <-w.Events():
		if len(ev.Diff.Opened) == 0 {
			t.Fatalf("expected opened ports, got none")
		}
		found := false
		for _, p := range ev.Diff.Opened {
			if p == port {
				found = true
			}
		}
		if !found {
			t.Fatalf("expected port %d in opened, got %v", port, ev.Diff.Opened)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for watcher event")
	}
}

func TestWatcherNoEventWhenNothingChanges(t *testing.T) {
	ln, port := openTCPListener(t)
	defer ln.Close()

	scanner := NewScanner(port, port, 50*time.Millisecond)
	baseline := NewBaseline("")
	// seed baseline with the port already open
	baseline.Set([]int{port})

	w := NewWatcher(scanner, nil, baseline, 30*time.Millisecond)
	w.Start()
	defer w.Stop()

	select {
	case ev := <-w.Events():
		t.Fatalf("unexpected event: %+v", ev)
	case <-time.After(200 * time.Millisecond):
		// expected: no event
	}
}

func TestWatcherSetsBaselineOnFirstScan(t *testing.T) {
	ln, port := openTCPListener(t)
	defer ln.Close()

	scanner := NewScanner(port, port, 50*time.Millisecond)
	baseline := NewBaseline("")
	// do NOT seed baseline — first scan should set it without emitting event

	w := NewWatcher(scanner, nil, baseline, 30*time.Millisecond)
	w.Start()
	defer w.Stop()

	time.Sleep(150 * time.Millisecond)

	select {
	case ev := <-w.Events():
		t.Fatalf("unexpected event on first scan: %+v", ev)
	default:
	}

	if got := baseline.Get(); len(got) == 0 {
		t.Fatal("expected baseline to be populated after first scan")
	}
}
