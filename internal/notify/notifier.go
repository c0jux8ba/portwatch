package notify

import "github.com/user/portwatch/internal/ports"

// Notifier is the common interface for all alert back-ends.
type Notifier interface {
	Notify(diff ports.Diff) error
}

// multi fans out to multiple notifiers.
type multi struct {
	notifiers []Notifier
}

// NewMulti returns a Notifier that forwards to all provided notifiers.
// Errors from individual notifiers are collected but do not stop others.
func NewMulti(nn ...Notifier) Notifier {
	filtered := make([]Notifier, 0, len(nn))
	for _, n := range nn {
		if n != nil {
			filtered = append(filtered, n)
		}
	}
	return &multi{notifiers: filtered}
}

func (m *multi) Notify(diff ports.Diff) error {
	var last error
	for _, n := range m.notifiers {
		if err := n.Notify(diff); err != nil {
			last = err
		}
	}
	return last
}

// noop is a no-operation notifier used when no back-ends are configured.
type noop struct{}

func (noop) Notify(_ ports.Diff) error { return nil }

// NewNoop returns a Notifier that silently discards all events.
func NewNoop() Notifier { return noop{} }
