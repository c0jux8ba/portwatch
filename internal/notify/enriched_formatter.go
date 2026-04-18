package notify

import (
	"fmt"
	"strings"

	"github.com/user/portwatch/internal/ports"
)

// EnrichedFormatter formats diff output with service and process annotations.
type EnrichedFormatter struct {
	enricher *ports.Enricher
	hostname string
}

// NewEnrichedFormatter creates an EnrichedFormatter.
func NewEnrichedFormatter(e *ports.Enricher, hostname string) *EnrichedFormatter {
	if hostname == "" {
		hostname = "localhost"
	}
	return &EnrichedFormatter{enricher: e, hostname: hostname}
}

// Format renders a human-readable string from a Diff using enriched port metadata.
func (f *EnrichedFormatter) Format(d ports.Diff) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[portwatch] %s\n", f.hostname))

	if len(d.Opened) > 0 {
		sb.WriteString("  Opened:\n")
		for _, info := range f.enricher.Enrich(d.Opened) {
			sb.WriteString(formatLine(info))
		}
	}
	if len(d.Closed) > 0 {
		sb.WriteString("  Closed:\n")
		for _, info := range f.enricher.Enrich(d.Closed) {
			sb.WriteString(formatLine(info))
		}
	}
	return strings.TrimRight(sb.String(), "\n")
}

// HasChanges reports whether the diff contains any opened or closed ports.
func (f *EnrichedFormatter) HasChanges(d ports.Diff) bool {
	return len(d.Opened) > 0 || len(d.Closed) > 0
}

func formatLine(info ports.PortInfo) string {
	base := fmt.Sprintf("    :%d (%s)", info.Port, info.Service)
	if info.Process != "" {
		base += fmt.Sprintf(" — %s [%d]", info.Process, info.PID)
	}
	return base + "\n"
}
