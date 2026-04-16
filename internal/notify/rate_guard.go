package notify

import (
	"strings"
	"sync"
	"time"

	"github.com/userwatch/internal/ports"
)

// RateGuard wraps a Notifier and suppresses repeated notifications
// for the same diff within a configurable cooldown window.
type RateGuard struct {
	inner    Notifier
	cooldown time.Duration
	clock    func() time.Time

	mu      sync.Mutex
	lastKey string
	lastAt  time.Time
}

// NewRateGuard returns a RateGuard that delegates to inner and suppresses
// duplicate notifications within cooldown.
func NewRateGuard(inner Notifier, cooldown time.Duration) *RateGuard {
	return &RateGuard{
		inner:    inner,
		cooldown: cooldown,
		clock:    time.Now,
	}
}

func newRateGuardWithClock(inner Notifier, cooldown time.Duration, clock func() time.Time) *RateGuard {
	return &RateGuard{
		inner:    inner,
		cooldown: cooldown,
		clock:    clock,
	}
}

// Notify forwards the diff to the inner notifier unless an identical diff
// was already forwarded within the cooldown window.
func (g *RateGuard) Notify(d ports.Diff) error {
	if d.IsEmpty() {
		return nil
	}

	key := diffKey(d)
	now := g.clock()

	g.mu.Lock()
	defer g.mu.Unlock()

	if key == g.lastKey && now.Sub(g.lastAt) < g.cooldown {
		return nil
	}

	g.lastKey = key
	g.lastAt = now
	return g.inner.Notify(d)
}

// diffKey produces a stable string key representing the diff content.
func diffKey(d ports.Diff) string {
	var b strings.Builder
	b.WriteString("o:")
	for _, p := range d.Opened {
		b.WriteString(intsToStrings([]int{p})[0])
		b.WriteByte(',')
	}
	b.WriteString("c:")
	for _, p := range d.Closed {
		b.WriteString(intsToStrings([]int{p})[0])
		b.WriteByte(',')
	}
	return b.String()
}
