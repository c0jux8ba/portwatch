package notify

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/portwatch/internal/ports"
)

// ConsoleNotifier writes port change events to a writer (default: stdout)
// with timestamps and colour-coded output.
type ConsoleNotifier struct {
	out    io.Writer
	prefix string
}

func NewConsoleNotifier(prefix string, out io.Writer) *ConsoleNotifier {
	if out == nil {
		out = os.Stdout
	}
	if prefix == "" {
		prefix = "portwatch"
	}
	return &ConsoleNotifier{out: out, prefix: prefix}
}

func (c *ConsoleNotifier) Notify(d ports.Diff) error {
	if d.IsEmpty() {
		return nil
	}
	ts := time.Now().Format(time.RFC3339)
	for _, p := range d.Opened {
		fmt.Fprintf(c.out, "[%s] %s \033[32mOPENED\033[0m port %d\n", ts, c.prefix, p)
	}
	for _, p := range d.Closed {
		fmt.Fprintf(c.out, "[%s] %s \033[31mCLOSED\033[0m port %d\n", ts, c.prefix, p)
	}
	return nil
}
