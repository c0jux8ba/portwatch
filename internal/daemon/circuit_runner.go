package daemon

import (
	"log"
	"time"

	"github.com/user/portwatch/internal/notify"
	"github.com/user/portwatch/internal/ports"
)

// CircuitRunner wraps a Runner and applies a circuit breaker to its notifier
// so that a flapping downstream does not spam retries indefinitely.
type CircuitRunner struct {
	inner   Runner
	breaker *notify.CircuitBreaker
}

// Runner is the minimal interface used by CircuitRunner.
type Runner interface {
	Tick() error
}

// NewCircuitRunner wraps runner with a circuit breaker that opens after
// maxFailures consecutive errors and resets after cooldown.
func NewCircuitRunner(inner Runner, n notify.Notifier, maxFailures int, cooldown time.Duration) *CircuitRunner {
	return &CircuitRunner{
		inner:   inner,
		breaker: notify.NewCircuitBreaker(n, maxFailures, cooldown),
	}
}

func (cr *CircuitRunner) Tick() error {
	err := cr.inner.Tick()
	if err != nil {
		log.Printf("[circuit_runner] tick error: %v", err)
	}
	return err
}

func (cr *CircuitRunner) NotifyDiff(d ports.Diff) error {
	return cr.breaker.Notify(d)
}
