package daemon

import (
	"fmt"

	"github.com/user/portwatch/internal/notify"
	"github.com/user/portwatch/internal/ports"
)

// MattermostRunner wires a Mattermost notifier into the daemon tick loop.
type MattermostRunner struct {
	inner  Runner
	notify *notify.MattermostNotifier
}

// NewMattermostRunner returns a Runner that posts port-change events to
// Mattermost. webhookURL is required; username and channel are optional.
// Returns an error if webhookURL is empty.
func NewMattermostRunner(inner Runner, webhookURL, username, channel string) (*MattermostRunner, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("mattermost runner: webhookURL must not be empty")
	}
	return &MattermostRunner{
		inner:  inner,
		notify: notify.NewMattermostNotifier(webhookURL, username, channel),
	}, nil
}

// Tick runs the inner runner and, on a non-empty diff, posts to Mattermost.
func (m *MattermostRunner) Tick() (ports.Diff, error) {
	diff, err := m.inner.Tick()
	if err != nil {
		return diff, err
	}
	if diff.IsEmpty() {
		return diff, nil
	}
	if notifyErr := m.notify.Notify(diff); notifyErr != nil {
		return diff, fmt.Errorf("mattermost runner: notify: %w", notifyErr)
	}
	return diff, nil
}
