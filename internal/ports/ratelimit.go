package ports

import (
	"sync"
	"time"
)

// RateLimiter suppresses repeated alerts for the same port within a cooldown window.
type RateLimiter struct {
	mu       sync.Mutex
	cooldown time.Duration
	last     map[int]time.Time
	now      func() time.Time
}

// NewRateLimiter creates a RateLimiter with the given cooldown duration.
func NewRateLimiter(cooldown time.Duration) *RateLimiter {
	return &RateLimiter{
		cooldown: cooldown,
		last:     make(map[int]time.Time),
		now:      time.Now,
	}
}

// Allow returns true if the port has not been seen within the cooldown window.
// If allowed, it records the current time for that port.
func (r *RateLimiter) Allow(port int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.now()
	if last, ok := r.last[port]; ok {
		if now.Sub(last) < r.cooldown {
			return false
		}
	}
	r.last[port] = now
	return true
}

// FilterDiff removes ports from a Diff that are within the cooldown window,
// returning a new Diff containing only the ports that are allowed through.
func (r *RateLimiter) FilterDiff(d Diff) Diff {
	opened := make([]int, 0, len(d.Opened))
	for _, p := range d.Opened {
		if r.Allow(p) {
			opened = append(opened, p)
		}
	}
	closed := make([]int, 0, len(d.Closed))
	for _, p := range d.Closed {
		if r.Allow(p) {
			closed = append(closed, p)
		}
	}
	return Diff{Opened: opened, Closed: closed}
}
