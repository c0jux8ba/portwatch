package notify

import (
	"io"
	"encoding/json"
	"time"

	"github.com/user/portwatch/internal/ports"
)

// TeeNotifier forwards every notification to a wrapped Notifier and also
// writes a JSON record of the diff to an io.Writer (e.g. a log file).
type TeeNotifier struct {
	wrapped  Notifier
	w        io.Writer
}

type teeRecord struct {
	Timestamp string   `json:"timestamp"`
	Opened    []int    `json:"opened,omitempty"`
	Closed    []int    `json:"closed,omitempty"`
}

// NewTeeNotifier returns a TeeNotifier that delegates to n and writes JSON
// records to w. If w is nil, os.Stdout is NOT used — writes are simply skipped.
func NewTeeNotifier(n Notifier, w io.Writer) *TeeNotifier {
	return &TeeNotifier{wrapped: n, w: w}
}

func (t *TeeNotifier) Notify(d ports.Diff) error {
	if d.IsEmpty() {
		return nil
	}

	if t.w != nil {
		rec := teeRecord{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Opened:    d.Opened,
			Closed:    d.Closed,
		}
		if b, err := json.Marshal(rec); err == nil {
			_, _ = t.w.Write(append(b, '\n'))
		}
	}

	if t.wrapped != nil {
		return t.wrapped.Notify(d)
	}
	return nil
}
