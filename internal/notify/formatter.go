package notify

import (
	"fmt"
	"strings"

	"github.com/user/portwatch/internal/ports"
)

// Formatter enriches a Diff with human-readable service names using a
// ports.Resolver. It is used by notifiers to produce richer alert messages.
type Formatter struct {
	resolver *ports.Resolver
}

// NewFormatter creates a Formatter backed by the given Resolver.
// If resolver is nil a new default Resolver is created.
func NewFormatter(resolver *ports.Resolver) *Formatter {
	if resolver == nil {
		resolver = ports.NewResolver()
	}
	return &Formatter{resolver: resolver}
}

// FormatOpened returns a human-readable line for newly opened ports.
func (f *Formatter) FormatOpened(diff ports.Diff) string {
	if len(diff.Opened) == 0 {
		return ""
	}
	names := f.resolver.LookupAll(diff.Opened)
	return fmt.Sprintf("opened: %s", strings.Join(names, ", "))
}

// FormatClosed returns a human-readable line for newly closed ports.
func (f *Formatter) FormatClosed(diff ports.Diff) string {
	if len(diff.Closed) == 0 {
		return ""
	}
	names := f.resolver.LookupAll(diff.Closed)
	return fmt.Sprintf("closed: %s", strings.Join(names, ", "))
}

// FormatSummary returns a multi-line summary of all changes in the diff.
func (f *Formatter) FormatSummary(diff ports.Diff) string {
	var parts []string
	if line := f.FormatOpened(diff); line != "" {
		parts = append(parts, line)
	}
	if line := f.FormatClosed(diff); line != "" {
		parts = append(parts, line)
	}
	if len(parts) == 0 {
		return "no changes"
	}
	return strings.Join(parts, "\n")
}
