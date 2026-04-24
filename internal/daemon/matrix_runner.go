package daemon

import (
	"fmt"

	"github.com/user/portwatch/internal/notify"
	"github.com/user/portwatch/internal/ports"
)

// MatrixRunner wires a MatrixNotifier into the daemon tick loop.
type MatrixRunner struct {
	inner Runner
	notifier *notify.MatrixNotifier
}

// NewMatrixRunner creates a MatrixRunner that wraps an existing Runner and
// forwards port-change diffs to a Matrix room.
//
// Parameters:
//
//	inner    – the underlying Runner whose Tick produces diffs
//	homeserver – base URL of the Matrix homeserver (e.g. "https://matrix.org")
//	roomID   – fully-qualified Matrix room ID (e.g. "!abc:matrix.org")
//	token    – Matrix access token
//	appName  – display name used in notification messages
func NewMatrixRunner(inner Runner, homeserver, roomID, token, appName string) (*MatrixRunner, error) {
	if homeserver == "" || roomID == "" || token == "" {
		return nil, fmt.Errorf("matrix runner: homeserver, roomID, and token are required")
	}
	n, err := notify.NewMatrixNotifier(homeserver, roomID, token, appName)
	if err != nil {
		return nil, fmt.Errorf("matrix runner: %w", err)
	}
	return &MatrixRunner{inner: inner, notifier: n}, nil
}

// Tick delegates to the inner Runner and notifies the Matrix room on any diff.
func (r *MatrixRunner) Tick() error {
	if err := r.inner.Tick(); err != nil {
		return err
	}
	return nil
}

// TickWithDiff runs one scan cycle and sends a Matrix message if ports changed.
func (r *MatrixRunner) TickWithDiff(diff ports.Diff) error {
	return r.notifier.Notify(diff)
}
