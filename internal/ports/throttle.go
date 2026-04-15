package ports

import (
	"sync"
	"time"
)

// Throttle limits how frequently a scan can be triggered,
// ensuring a minimum interval between consecutive scans.
type Throttle struct {
	mu       sync.Mutex
	last     time.Time
	minGap   time.Duration
	clock    func() time.Time
}

// NewThrottle creates a Throttle that enforces at least minGap between scans.
func NewThrottle(minGap time.Duration) *Throttle {
	return &Throttle{
		minGap: minGap,
		clock:  time.Now,
	}
}

// newThrottleWithClock is used in tests to inject a custom clock.
func newThrottleWithClock(minGap time.Duration, clock func() time.Time) *Throttle {
	return &Throttle{
		minGap: minGap,
		clock:  clock,
	}
}

// Allow returns true if enough time has passed since the last allowed call.
// If allowed, it records the current time as the last scan time.
func (t *Throttle) Allow() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.clock()
	if t.last.IsZero() || now.Sub(t.last) >= t.minGap {
		t.last = now
		return true
	}
	return false
}

// Reset clears the last scan timestamp, allowing the next call immediately.
func (t *Throttle) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.last = time.Time{}
}

// LastScan returns the time of the last allowed scan (zero if never).
func (t *Throttle) LastScan() time.Time {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.last
}
