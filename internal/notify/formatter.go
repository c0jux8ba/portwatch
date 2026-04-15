package notify

import (
	"fmt"
	"os"
	"strings"

	"github.com/user/portwatch/internal/ports"
)

// Formatter builds human-readable alert messages from a port Diff.
type Formatter struct {
	hostname string
}

// NewFormatter creates a Formatter. If hostname is empty, the system
// hostname is used as a fallback.
func NewFormatter(hostname string) *Formatter {
	if hostname == "" {
		if h, err := os.Hostname(); err == nil {
			hostname = h
		} else {
			hostname = "localhost"
		}
	}
	return &Formatter{hostname: hostname}
}

// Format returns a formatted alert string for the given Diff.
// Returns an empty string when the diff has no changes.
func (f *Formatter) Format(diff ports.Diff) string {
	if diff.IsEmpty() {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[portwatch] Port change detected on %s\n", f.hostname))

	if len(diff.Opened) > 0 {
		sb.WriteString(fmt.Sprintf("  Opened: %s\n", joinInts(diff.Opened)))
	}
	if len(diff.Closed) > 0 {
		sb.WriteString(fmt.Sprintf("  Closed: %s\n", joinInts(diff.Closed)))
	}

	return strings.TrimRight(sb.String(), "\n")
}

// Subject returns a short one-line summary suitable for webhook titles
// or notification headings.
func (f *Formatter) Subject(diff ports.Diff) string {
	parts := []string{}
	if len(diff.Opened) > 0 {
		parts = append(parts, fmt.Sprintf("%d opened", len(diff.Opened)))
	}
	if len(diff.Closed) > 0 {
		parts = append(parts, fmt.Sprintf("%d closed", len(diff.Closed)))
	}
	if len(parts) == 0 {
		return ""
	}
	return fmt.Sprintf("portwatch [%s]: %s", f.hostname, strings.Join(parts, ", "))
}

func joinInts(vals []int) string {
	ss := make([]string, len(vals))
	for i, v := range vals {
		ss[i] = fmt.Sprintf("%d", v)
	}
	return strings.Join(ss, ", ")
}
