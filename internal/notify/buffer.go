package notify

import (
	"sync"

	"github.com/user/portwatch/internal/ports"
)

// BufferedNotifier accumulates diffs and flushes them as a single
// combined notification when Flush is called.
type BufferedNotifier struct {
	mu      sync.Mutex
	opened  []int
	closed  []int
	delegate Notifier
}

// NewBufferedNotifier wraps delegate, batching diffs until Flush.
func NewBufferedNotifier(delegate Notifier) *BufferedNotifier {
	return &BufferedNotifier{delegate: delegate}
}

// Notify accumulates the diff without forwarding it yet.
func (b *BufferedNotifier) Notify(d ports.Diff) error {
	if d.IsEmpty() {
		return nil
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.opened = append(b.opened, d.Opened...)
	b.closed = append(b.closed, d.Closed...)
	return nil
}

// Flush sends all accumulated changes as one diff and resets the buffer.
func (b *BufferedNotifier) Flush() error {
	b.mu.Lock()
	opened := b.opened
	closed := b.closed
	b.opened = nil
	b.closed = nil
	b.mu.Unlock()

	combined := ports.Diff{Opened: opened, Closed: closed}
	if combined.IsEmpty() {
		return nil
	}
	return b.delegate.Notify(combined)
}

// Len returns the number of buffered port events.
func (b *BufferedNotifier) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.opened) + len(b.closed)
}
