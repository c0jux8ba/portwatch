package ports

import (
	"strings"
	"testing"
)

func TestParseLSOFOutputValidLine(t *testing.T) {
	input := `COMMAND   PID USER   FD   TYPE DEVICE SIZE/OFF NODE NAME
nginx    1234 root    6u  IPv4  12345      0t0  TCP *:80 (LISTEN)
`
	info := parseLSOFOutput(80, input)
	if info == nil {
		t.Fatal("expected ProcessInfo, got nil")
	}
	if info.PID != 1234 {
		t.Errorf("expected PID 1234, got %d", info.PID)
	}
	if info.Name != "nginx" {
		t.Errorf("expected name nginx, got %s", info.Name)
	}
	if info.Port != 80 {
		t.Errorf("expected port 80, got %d", info.Port)
	}
}

func TestParseLSOFOutputHeaderOnly(t *testing.T) {
	input := "COMMAND   PID USER   FD   TYPE DEVICE SIZE/OFF NODE NAME\n"
	info := parseLSOFOutput(80, input)
	if info != nil {
		t.Errorf("expected nil for header-only output, got %+v", info)
	}
}

func TestParseLSOFOutputEmpty(t *testing.T) {
	info := parseLSOFOutput(8080, "")
	if info != nil {
		t.Errorf("expected nil for empty output, got %+v", info)
	}
}

func TestParseLSOFOutputInvalidPID(t *testing.T) {
	input := "sshd   notapid root  3u IPv4 999 0t0 TCP *:22 (LISTEN)\n"
	info := parseLSOFOutput(22, input)
	if info != nil {
		t.Errorf("expected nil when PID is not numeric, got %+v", info)
	}
}

func TestParseLSOFOutputMultipleLines(t *testing.T) {
	lines := []string{
		"COMMAND   PID USER   FD   TYPE DEVICE SIZE/OFF NODE NAME",
		"python3  5678 user    4u  IPv4  99999      0t0  TCP *:5000 (LISTEN)",
		"node     9999 user    6u  IPv4  88888      0t0  TCP *:5000 (LISTEN)",
	}
	input := strings.Join(lines, "\n") + "\n"
	info := parseLSOFOutput(5000, input)
	if info == nil {
		t.Fatal("expected ProcessInfo, got nil")
	}
	// Should return the first non-header match
	if info.PID != 5678 {
		t.Errorf("expected PID 5678, got %d", info.PID)
	}
	if info.Name != "python3" {
		t.Errorf("expected name python3, got %s", info.Name)
	}
}

func TestNewProcessResolver(t *testing.T) {
	r := NewProcessResolver()
	if r == nil {
		t.Fatal("expected non-nil ProcessResolver")
	}
}
