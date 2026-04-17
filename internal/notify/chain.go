package notify

import "github.com/user/portwatch/internal/ports"

// ChainNotifier applies a series of middleware-style wrappers around a base Notifier.
// Each wrapper is a func(Notifier) Notifier, applied innermost-first.
type ChainNotifier struct {
	notifier Notifier
}

// NewChain builds a ChainNotifier by wrapping base with each middleware in order.
// The first middleware in the slice is the outermost wrapper.
func NewChain(base Notifier, middlewares ...func(Notifier) Notifier) *ChainNotifier {
	n := base
	for i := len(middlewares) - 1; i >= 0; i-- {
		n = middlewares[i](n)
	}
	return &ChainNotifier{notifier: n}
}

func (c *ChainNotifier) Notify(d ports.Diff) error {
	return c.notifier.Notify(d)
}
