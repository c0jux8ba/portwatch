package ports

import (
	"net"
	"testing"
	"time"
)

// startTestListener opens a TCP listener on an ephemeral port and returns it.
func startTestListener(t *testing.T) (net.Listener, int) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test listener: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	return ln, port
}

func TestScanDetectsOpenPort(t *testing.T) {
	ln, port := startTestListener(t)
	defer ln.Close()

	s := &Scanner{
		StartPort: port,
		EndPort:   port,
		Protocol:  "tcp",
		Timeout:   200 * time.Millisecond,
	}

	snap, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(snap.Ports) != 1 {
		t.Fatalf("expected 1 open port, got %d", len(snap.Ports))
	}
	if snap.Ports[0].Port != port {
		t.Errorf("expected port %d, got %d", port, snap.Ports[0].Port)
	}
	if snap.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestScanClosedPort(t *testing.T) {
	// Use a listener just to grab a free port, then close it immediately.
	ln, port := startTestListener(t)
	ln.Close()

	s := NewScanner(port, port)
	snap, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(snap.Ports) != 0 {
		t.Errorf("expected 0 open ports on closed port, got %d", len(snap.Ports))
	}
}

func TestScanInvalidRange(t *testing.T) {
	s := NewScanner(9000, 8000) // inverted range
	_, err := s.Scan()
	if err == nil {
		t.Error("expected error for invalid port range, got nil")
	}
}
