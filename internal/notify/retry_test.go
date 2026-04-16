package notify

import (
	"errors"
	"testing"
	"time"

	"github.com/user/portwatch/internal/ports"
)

type countingNotifier struct {
	calls  int
	failN  int // fail first N calls
	err    error
}

func (c *countingNotifier) Notify(_ ports.Diff) error {
	c.calls++
	if c.calls <= c.failN {
		return c.err
	}
	return nil
}

func noSleep(_ time.Duration) {}

func TestRetrySucceedsOnFirstAttempt(t *testing.T) {
	inner := &countingNotifier{}
	r := NewRetryNotifier(inner, 3, 0)
	r.sleep = noSleep
	if err := r.Notify(ports.Diff{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inner.calls != 1 {
		t.Fatalf("expected 1 call, got %d", inner.calls)
	}
}

func TestRetryRetriesOnFailure(t *testing.T) {
	inner := &countingNotifier{failN: 2, err: errors.New("boom")}
	r := NewRetryNotifier(inner, 3, 0)
	r.sleep = noSleep
	if err := r.Notify(ports.Diff{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inner.calls != 3 {
		t.Fatalf("expected 3 calls, got %d", inner.calls)
	}
}

func TestRetryReturnsErrorAfterExhaustion(t *testing.T) {
	inner := &countingNotifier{failN: 5, err: errors.New("always fails")}
	r := NewRetryNotifier(inner, 3, 0)
	r.sleep = noSleep
	err := r.Notify(ports.Diff{})
	if err == nil {
		t.Fatal("expected error")
	}
	if inner.calls != 3 {
		t.Fatalf("expected 3 calls, got %d", inner.calls)
	}
}

func TestRetrySleesBetweenAttempts(t *testing.T) {
	inner := &countingNotifier{failN: 2, err: errors.New("boom")}
	slept := 0
	r := NewRetryNotifier(inner, 3, 50*time.Millisecond)
	r.sleep = func(_ time.Duration) { slept++ }
	_ = r.Notify(ports.Diff{})
	if slept != 2 {
		t.Fatalf("expected 2 sleeps, got %d", slept)
	}
}
