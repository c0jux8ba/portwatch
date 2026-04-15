package daemon

import (
	"context"
	"log"
	"time"

	"github.com/user/portwatch/internal/notify"
	"github.com/user/portwatch/internal/ports"
)

// Runner executes one scan-diff-notify cycle and is separated from
// the ticker logic so it can be unit-tested independently.
type Runner struct {
	scanner  *ports.Scanner
	notifier notify.Notifier
	snapshot *ports.Snapshot
}

// NewRunner creates a Runner with an already-initialised snapshot (baseline).
func NewRunner(scanner *ports.Scanner, notifier notify.Notifier, snapshot *ports.Snapshot) *Runner {
	return &Runner{
		scanner:  scanner,
		notifier: notifier,
		snapshot: snapshot,
	}
}

// Tick performs a single scan iteration.
// It returns true if a change was detected and notifications were sent.
func (r *Runner) Tick(ctx context.Context) (bool, error) {
	current, err := r.scanner.Scan(ctx)
	if err != nil {
		return false, err
	}

	diff := ports.Compare(r.snapshot.Ports(), current)
	if diff.IsEmpty() {
		return false, nil
	}

	log.Printf("change detected — opened: %v  closed: %v", diff.Opened, diff.Closed)

	if err := r.notifier.Notify(diff); err != nil {
		log.Printf("notification error: %v", err)
	}

	r.snapshot.Update(current)
	return true, nil
}

// RunLoop blocks, calling Tick on every interval tick until ctx is cancelled.
func (r *Runner) RunLoop(ctx context.Context, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if _, err := r.Tick(ctx); err != nil {
				log.Printf("scan error: %v", err)
			}
		}
	}
}
