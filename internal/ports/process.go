package ports

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// ProcessInfo holds information about a process listening on a port.
type ProcessInfo struct {
	Port int
	PID  int
	Name string
}

// ProcessResolver resolves which process owns each open port via lsof.
type ProcessResolver struct {
	runner func() (string, error)
}

// NewProcessResolver returns a ProcessResolver that calls lsof.
func NewProcessResolver() *ProcessResolver {
	return &ProcessResolver{runner: runLSOF}
}

// Resolve returns a slice of ProcessInfo for all listening TCP ports.
func (p *ProcessResolver) Resolve() ([]ProcessInfo, error) {
	out, err := p.runner()
	if err != nil {
		return nil, fmt.Errorf("lsof: %w", err)
	}
	return parseLSOFOutput(out), nil
}

func runLSOF() (string, error) {
	out, err := exec.Command("lsof", "-iTCP", "-sTCP:LISTEN", "-Pn").Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func parseLSOFOutput(output string) []ProcessInfo {
	var result []ProcessInfo
	lines := strings.Split(output, "\n")
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}
		name := fields[0]
		pid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		addr := fields[8]
		colon := strings.LastIndex(addr, ":")
		if colon < 0 {
			continue
		}
		port, err := strconv.Atoi(addr[colon+1:])
		if err != nil {
			continue
		}
		result = append(result, ProcessInfo{Port: port, PID: pid, Name: name})
	}
	return result
}
