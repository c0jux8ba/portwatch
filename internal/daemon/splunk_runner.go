package daemon

import (
	"log"

	"github.com/user/portwatch/internal/notify"
	"github.com/user/portwatch/internal/ports"
)

// SplunkRunner wraps a Runner and forwards diffs to a Splunk HEC endpoint.
type SplunkRunner struct {
	inner   Runner
	notifier *notify.SplunkNotifier
}

// NewSplunkRunner creates a SplunkRunner that delegates scanning to inner and
// ships every non-empty diff to the provided Splunk notifier.
func NewSplunkRunner(inner Runner, endpoint, token string) *SplunkRunner {
	return &SplunkRunner{
		inner:    inner,
		notifier: notify.NewSplunkNotifier(endpoint, token),
	}
}

func (s *SplunkRunner) Tick() (ports.Diff, error) {
	diff, err := s.inner.Tick()
	if err != nil {
		return diff, err
	}
	if diff.IsEmpty() {
		return diff, nil
	}
	if nerr := s.notifier.Notify(diff); nerr != nil {
		log.Printf("splunk_runner: notify error: %v", nerr)
	}
	return diff, nil
}
