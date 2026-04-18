package notify

import (
	"errors"
	"sync"
	"time"

	"github.com/user/portwatch/internal/ports"
)

// ErrCircuitOpen is returned when the circuit breaker is open.
var ErrCircuitOpen = errors.New("circuit breaker open: too many consecutive failures")

// CircuitBreaker wraps a Notifier and stops forwarding after too many failures.
type CircuitBreaker struct {
	mu           sync.Mutex
	wrapped      Notifier
	maxFailures  int
	cooldown     time.Duration
	failures     int
	openUntil    time.Time
	now          func() time.Time
}

// NewCircuitBreaker returns a Notifier that opens after maxFailures consecutive
// errors and recovers after cooldown.
func NewCircuitBreaker(n Notifier, maxFailures int, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		wrapped:     n,
		maxFailures: maxFailures,
		cooldown:    cooldown,
		now:         time.Now,
	}
}

func (cb *CircuitBreaker) Notify(d ports.Diff) error {
	if d.IsEmpty() {
		return nil
	}
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.now().Before(cb.openUntil) {
		return ErrCircuitOpen
	}

	err := cb.wrapped.Notify(d)
	if err != nil {
		cb.failures++
		if cb.failures >= cb.maxFailures {
			cb.openUntil = cb.now().Add(cb.cooldown)
			cb.failures = 0
		}
		return err
	}
	cb.failures = 0
	return nil
}
