package notify

import (
	"fmt"

	"github.com/user/portwatch/internal/ports"
)

// Notifier is the common interface for all alerting backends.
type Notifier interface {
	Notify(diff ports.Diff) error
}

// Multi fans a single Diff out to multiple Notifier implementations.
// Errors from individual notifiers are collected and returned together.
type Multi struct {
	notifiers []Notifier
}

// NewMulti creates a Multi notifier wrapping the provided notifiers.
func NewMulti(notifiers ...Notifier) *Multi {
	return &Multi{notifiers: notifiers}
}

// Notify calls every registered notifier and collects errors.
func (m *Multi) Notify(diff ports.Diff) error {
	var errs []error
	for _, n := range m.notifiers {
		if err := n.Notify(diff); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return &MultiError{Errors: errs}
}

// MultiError aggregates errors from multiple notifiers.
type MultiError struct {
	Errors []error
}

func (e *MultiError) Error() string {
	msg := fmt.Sprintf("%d notifier(s) failed:", len(e.Errors))
	for i, err := range e.Errors {
		msg += fmt.Sprintf(" [%d] %s", i+1, err.Error())
	}
	return msg
}
