package notify

import (
	"errors"
	"testing"
	"time"

	"github.com/user/portwatch/internal/ports"
)

var errBoom = errors.New("boom")

type countingNotifier struct {
	calls int
	err   error
}

func (c *countingNotifier) Notify(_ ports.Diff) error {
	c.calls++
	return c.err
}

func fixedNow(t time.Time) func() time.Time { return func() time.Time { return t } }

func TestCircuitBreakerAllowsOnSuccess(t *testing.T) {
	n := &countingNotifier{}
	cb := NewCircuitBreaker(n, 3, time.Minute)
	if err := cb.Notify(sampleDiff()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.calls != 1 {
		t.Fatalf("expected 1 call, got %d", n.calls)
	}
}

func TestCircuitBreakerOpensAfterMaxFailures(t *testing.T) {
	n := &countingNotifier{err: errBoom}
	cb := NewCircuitBreaker(n, 2, time.Minute)
	cb.Notify(sampleDiff())
	cb.Notify(sampleDiff())
	err := cb.Notify(sampleDiff())
	if !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
	if n.calls != 2 {
		t.Fatalf("expected 2 calls before open, got %d", n.calls)
	}
}

func TestCircuitBreakerRecoversAfterCooldown(t *testing.T) {
	n := &countingNotifier{err: errBoom}
	base := time.Now()
	cb := NewCircuitBreaker(n, 1, time.Minute)
	cb.now = fixedNow(base)
	cb.Notify(sampleDiff()) // triggers open

	// still open
	cb.now = fixedNow(base.Add(30 * time.Second))
	if err := cb.Notify(sampleDiff()); !errors.Is(err, ErrCircuitOpen) {
		t.Fatal("expected open circuit")
	}

	// after cooldown, allow again
	n.err = nil
	cb.now = fixedNow(base.Add(2 * time.Minute))
	if err := cb.Notify(sampleDiff()); err != nil {
		t.Fatalf("expected recovery, got %v", err)
	}
}

func TestCircuitBreakerSkipsEmptyDiff(t *testing.T) {
	n := &countingNotifier{}
	cb := NewCircuitBreaker(n, 3, time.Minute)
	cb.Notify(ports.Diff{})
	if n.calls != 0 {
		t.Fatal("should skip empty diff")
	}
}

func TestCircuitBreakerResetsFailuresOnSuccess(t *testing.T) {
	n := &countingNotifier{err: errBoom}
	cb := NewCircuitBreaker(n, 3, time.Minute)
	cb.Notify(sampleDiff())
	cb.Notify(sampleDiff())
	n.err = nil
	cb.Notify(sampleDiff()) // success resets counter
	n.err = errBoom
	cb.Notify(sampleDiff())
	cb.Notify(sampleDiff())
	// should open again only after 3 more failures from reset
	if errors.Is(cb.Notify(sampleDiff()), ErrCircuitOpen) {
		t.Fatal("should not be open yet after reset")
	}
}
