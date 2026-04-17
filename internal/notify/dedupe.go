package notify

import (
	"sync"

	"github.com/user/portwatch/internal/ports"
)

// DedupeNotifier wraps a Notifier and suppresses consecutive identical diffs.
type DedupeNotifier struct {
	inner   Notifier
	mu      sync.Mutex
	lastKey string
}

// NewDedupeNotifier returns a Notifier that skips a notification when the
// diff is identical to the previous one that was actually sent.
func NewDedupeNotifier(inner Notifier) *DedupeNotifier {
	return &DedupeNotifier{inner: inner}
}

func (d *DedupeNotifier) Notify(diff ports.Diff) error {
	if diff.IsEmpty() {
		return nil
	}

	key := diffKey(diff)

	d.mu.Lock()
	if key == d.lastKey {
		d.mu.Unlock()
		return nil
	}
	d.lastKey = key
	d.mu.Unlock()

	return d.inner.Notify(diff)
}
