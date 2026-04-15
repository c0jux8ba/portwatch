package ports

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// ProcessInfo holds metadata about the process bound to a port.
type ProcessInfo struct {
	Port int
	PID  int
	Name string
}

// ProcessResolver looks up which process is listening on a given port.
type ProcessResolver struct{}

// NewProcessResolver creates a new ProcessResolver.
func NewProcessResolver() *ProcessResolver {
	return &ProcessResolver{}
}

// Lookup returns ProcessInfo for the given port, or nil if not found or unsupported.
func (r *ProcessResolver) Lookup(port int) *ProcessInfo {
	out, err := runLSOF(port)
	if err != nil {
		return nil
	}
	return parseLSOFOutput(port, out)
}

// runLSOF executes lsof for the given port and returns its output.
func runLSOF(port int) (string, error) {
	cmd := exec.Command("lsof", "-iTCP:"+strconv.Itoa(port), "-sTCP:LISTEN", "-n", "-P")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("lsof failed: %w", err)
	}
	return string(out), nil
}

// parseLSOFOutput parses lsof output and extracts the first matching process.
func parseLSOFOutput(port int, output string) *ProcessInfo {
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		// lsof columns: COMMAND PID USER FD TYPE DEVICE SIZE/OFF NODE NAME
		if len(fields) < 2 {
			continue
		}
		name := fields[0]
		if strings.EqualFold(name, "COMMAND") {
			continue // header
		}
		pid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		return &ProcessInfo{
			Port: port,
			PID:  pid,
			Name: name,
		}
	}
	return nil
}
