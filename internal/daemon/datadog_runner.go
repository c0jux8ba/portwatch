package daemon

import (
	"log"

	"github.com/user/portwatch/internal/notify"
	"github.com/user/portwatch/internal/ports"
)

// DatadogRunner wraps a Runner and forwards diffs to a DatadogNotifier.
type DatadogRunner struct {
	inner     Runner
	notifier  *notify.DatadogNotifier
}

type Runner interface {
	Tick() (ports.Diff, error)
}

func NewDatadogRunner(inner Runner, apiKey, endpoint string) *DatadogRunner {
	return &DatadogRunner{
		inner:    inner,
		notifier: notify.NewDatadogNotifier(apiKey, endpoint),
	}
}

func (d *DatadogRunner) Tick() (ports.Diff, error) {
	diff, err := d.inner.Tick()
	if err != nil {
		return diff, err
	}
	if !diff.IsEmpty() {
		if nerr := d.notifier.Notify(diff); nerr != nil {
			log.Printf("datadog runner: notify error: %v", nerr)
		}
	}
	return diff, nil
}
