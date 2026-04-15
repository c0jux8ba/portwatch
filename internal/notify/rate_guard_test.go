package notify

import (
	"errors"
	"testing"
	"time"

	"github.com/user/portwatch/internal/ports"
)

type countingNotifier struct {
	calls []ports.Diff
	err   error
}

func (c *countingNotifier) Notify(d ports.Diff) error {
	c.calls = append(c.calls, d)
	return c.err
}

func advancingClock(start time.Time, step time.Duration) func() time.Time {
	t := start
	return func() time.Time {
		now := t
		t = t.Add(step)
		return now
	}
}

func TestRateGuardAllowsFirstNotification(t *testing.T) {
	inner := &countingNotifier{}
	g := newRateGuardWithClock(inner, 5*time.Second, func() time.Time { return time.Now() })
	d := ports.Diff{Opened: []int{8080}}
	if err := g.Notify(d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(inner.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(inner.calls))
	}
}

func TestRateGuardSuppressesDuplicateWithinCooldown(t *testing.T) {
	inner := &countingNotifier{}
	now := time.Unix(1000, 0)
	clock := func() time.Time { return now }
	g := newRateGuardWithClock(inner, 10*time.Second, clock)
	d := ports.Diff{Opened: []int{9090}}
	_ = g.Notify(d)
	_ = g.Notify(d) // same diff, same time — should be suppressed
	if len(inner.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(inner.calls))
	}
}

func TestRateGuardAllowsAfterCooldown(t *testing.T) {
	inner := &countingNotifier{}
	base := time.Unix(2000, 0)
	cooldown := 5 * time.Second
	clock := advancingClock(base, cooldown+time.Second)
	g := newRateGuardWithClock(inner, cooldown, clock)
	d := ports.Diff{Opened: []int{3000}}
	_ = g.Notify(d)
	_ = g.Notify(d) // clock advanced beyond cooldown
	if len(inner.calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(inner.calls))
	}
}

func TestRateGuardAllowsDifferentDiff(t *testing.T) {
	inner := &countingNotifier{}
	now := time.Unix(3000, 0)
	clock := func() time.Time { return now }
	g := newRateGuardWithClock(inner, 30*time.Second, clock)
	_ = g.Notify(ports.Diff{Opened: []int{80}})
	_ = g.Notify(ports.Diff{Opened: []int{443}})
	if len(inner.calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(inner.calls))
	}
}

func TestRateGuardSkipsEmptyDiff(t *testing.T) {
	inner := &countingNotifier{}
	g := NewRateGuard(inner, time.Minute)
	_ = g.Notify(ports.Diff{})
	if len(inner.calls) != 0 {
		t.Fatalf("expected 0 calls, got %d", len(inner.calls))
	}
}

func TestRateGuardPropagatesError(t *testing.T) {
	want := errors.New("send failed")
	inner := &countingNotifier{err: want}
	g := NewRateGuard(inner, time.Minute)
	got := g.Notify(ports.Diff{Opened: []int{8000}})
	if got != want {
		t.Fatalf("expected error %v, got %v", want, got)
	}
}
