package notify

import (
	"fmt"
	"time"

	"github.com/user/portwatch/internal/ports"
)

// RetryNotifier wraps a Notifier and retries on failure.
type RetryNotifier struct {
	inner    Notifier
	attempts int
	delay    time.Duration
	sleep    func(time.Duration)
}

// NewRetryNotifier creates a RetryNotifier with the given attempt count and delay.
func NewRetryNotifier(inner Notifier, attempts int, delay time.Duration) *RetryNotifier {
	return &RetryNotifier{
		inner:    inner,
		attempts: attempts,
		delay:    delay,
		sleep:    time.Sleep,
	}
}

func (r *RetryNotifier) Notify(diff ports.Diff) error {
	if r.attempts <= 0 {
		return r.inner.Notify(diff)
	}
	var lastErr error
	for i := 0; i < r.attempts; i++ {
		if err := r.inner.Notify(diff); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if i < r.attempts-1 {
			r.sleep(r.delay)
		}
	}
	return fmt.Errorf("notify failed after %d attempts: %w", r.attempts, lastErr)
}
