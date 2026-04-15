package ports

import (
	"fmt"
	"net"
	"time"
)

// PortState represents the state of a single port.
type PortState struct {
	Port     int
	Protocol string
	Open     bool
	Address  string
}

// Snapshot holds all open ports at a point in time.
type Snapshot struct {
	Timestamp time.Time
	Ports     []PortState
}

// Scanner scans local open ports within a given range.
type Scanner struct {
	StartPort int
	EndPort   int
	Protocol  string
	Timeout   time.Duration
}

// NewScanner creates a Scanner with sensible defaults.
func NewScanner(start, end int) *Scanner {
	return &Scanner{
		StartPort: start,
		EndPort:   end,
		Protocol:  "tcp",
		Timeout:   500 * time.Millisecond,
	}
}

// Scan probes each port in the range and returns a Snapshot of open ports.
func (s *Scanner) Scan() (*Snapshot, error) {
	if s.StartPort < 1 || s.EndPort > 65535 || s.StartPort > s.EndPort {
		return nil, fmt.Errorf("invalid port range: %d-%d", s.StartPort, s.EndPort)
	}

	var open []PortState

	for port := s.StartPort; port <= s.EndPort; port++ {
		addr := fmt.Sprintf("127.0.0.1:%d", port)
		conn, err := net.DialTimeout(s.Protocol, addr, s.Timeout)
		if err == nil {
			conn.Close()
			open = append(open, PortState{
				Port:     port,
				Protocol: s.Protocol,
				Open:     true,
				Address:  addr,
			})
		}
	}

	return &Snapshot{
		Timestamp: time.Now(),
		Ports:     open,
	}, nil
}
