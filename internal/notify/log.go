package notify

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/portwatch/internal/ports"
)

// LogNotifier writes port change events to a writer (default: stdout).
type LogNotifier struct {
	out    io.Writer
	prefix string
}

// NewLogNotifier returns a LogNotifier writing to w.
// If w is nil, os.Stdout is used.
func NewLogNotifier(w io.Writer, prefix string) *LogNotifier {
	if w == nil {
		w = os.Stdout
	}
	if prefix == "" {
		prefix = "portwatch"
	}
	return &LogNotifier{out: w, prefix: prefix}
}

// Notify writes a human-readable log line for each change in diff.
func (l *LogNotifier) Notify(diff ports.Diff) error {
	if diff.IsEmpty() {
		return nil
	}
	ts := time.Now().Format(time.RFC3339)
	for _, p := range diff.Opened {
		fmt.Fprintf(l.out, "%s [%s] OPENED port %d\n", ts, l.prefix, p)
	}
	for _, p := range diff.Closed {
		fmt.Fprintf(l.out, "%s [%s] CLOSED port %d\n", ts, l.prefix, p)
	}
	return nil
}
