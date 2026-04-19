package daemon

import (
	"github.com/user/portwatch/internal/notify"
	"github.com/user/portwatch/internal/ports"
)

// NtfyRunner wraps a Runner and forwards diffs to an ntfy notifier.
type NtfyRunner struct {
	inner   Runner
	notifier *notify.NtfyNotifier
}

// NewNtfyRunner creates a NtfyRunner that wraps inner and notifies via ntfy.
func NewNtfyRunner(inner Runner, serverURL, topic string) *NtfyRunner {
	return &NtfyRunner{
		inner:    inner,
		notifier: notify.NewNtfyNotifier(serverURL, topic),
	}
}

func (r *NtfyRunner) Tick() (ports.Diff, error) {
	diff, err := r.inner.Tick()
	if err != nil {
		return diff, err
	}
	if notifyErr := r.notifier.Notify(diff); notifyErr != nil {
		return diff, notifyErr
	}
	return diff, nil
}
