package daemon

import (
	"log"
	"time"

	"github.com/user/portwatch/internal/ports"
)

// ThrottledRunner wraps a Runner with a Throttle so that rapid ticks
// (e.g. from a tight loop or signal-based trigger) do not cause excessive scans.
type ThrottledRunner struct {
	runner   *Runner
	throttle *ports.Throttle
}

// NewThrottledRunner creates a ThrottledRunner enforcing minGap between scans.
func NewThrottledRunner(r *Runner, minGap time.Duration) *ThrottledRunner {
	return &ThrottledRunner{
		runner:   r,
		throttle: ports.NewThrottle(minGap),
	}
}

// Tick performs a scan only if the throttle permits.
// Returns true if a scan was attempted, false if it was suppressed.
func (tr *ThrottledRunner) Tick() bool {
	if !tr.throttle.Allow() {
		log.Println("[throttle] scan suppressed — too soon since last scan")
		return false
	}
	if err := tr.runner.Tick(); err != nil {
		log.Printf("[throttle] scan error: %v", err)
	}
	return true
}

// Reset clears throttle state, useful for testing or forced re-scan.
func (tr *ThrottledRunner) Reset() {
	tr.throttle.Reset()
}

// LastScan returns the time of the most recent allowed scan.
func (tr *ThrottledRunner) LastScan() time.Time {
	return tr.throttle.LastScan()
}
