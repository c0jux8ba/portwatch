package notify

import "github.com/user/portwatch/internal/ports"

// FilterNotifier wraps a Notifier and only forwards diffs that pass a predicate.
type FilterNotifier struct {
	inner     Notifier
	predicate func(ports.Diff) bool
}

// NewFilterNotifier returns a FilterNotifier that calls inner only when predicate returns true.
func NewFilterNotifier(inner Notifier, predicate func(ports.Diff) bool) *FilterNotifier {
	return &FilterNotifier{inner: inner, predicate: predicate}
}

func (f *FilterNotifier) Notify(d ports.Diff) error {
	if d.IsEmpty() {
		return nil
	}
	if f.predicate != nil && !f.predicate(d) {
		return nil
	}
	return f.inner.Notify(d)
}
