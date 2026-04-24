package ports

import (
	"testing"
	"time"
)

func TestThrottleAllowsFirstCall(t *testing.T) {
	th := NewThrottle(5 * time.Second)
	if !th.Allow() {
		t.Fatal("expected first call to be allowed")
	}
}

func TestThrottleBlocksWithinGap(t *testing.T) {
	now := time.Now()
	th := newThrottleWithClock(5*time.Second, func() time.Time { return now })

	if !th.Allow() {
		t.Fatal("expected first call to be allowed")
	}
	// Same instant — within the gap
	if th.Allow() {
		t.Fatal("expected second call within gap to be blocked")
	}
}

func TestThrottleAllowsAfterGap(t *testing.T) {
	base := time.Now()
	calls := 0
	clock := func() time.Time {
		calls++
		if calls == 1 {
			return base
		}
		return base.Add(6 * time.Second)
	}
	th := newThrottleWithClock(5*time.Second, clock)

	if !th.Allow() {
		t.Fatal("expected first call to be allowed")
	}
	if !th.Allow() {
		t.Fatal("expected call after gap to be allowed")
	}
}

func TestThrottleReset(t *testing.T) {
	now := time.Now()
	th := newThrottleWithClock(5*time.Second, func() time.Time { return now })

	th.Allow() // record last
	if th.Allow() {
		t.Fatal("expected blocked before reset")
	}
	th.Reset()
	if !th.Allow() {
		t.Fatal("expected allowed after reset")
	}
}

func TestThrottleLastScanZeroInitially(t *testing.T) {
	th := NewThrottle(time.Second)
	if !th.LastScan().IsZero() {
		t.Fatal("expected LastScan to be zero before any call")
	}
}

func TestThrottleLastScanUpdated(t *testing.T) {
	now := time.Now()
	th := newThrottleWithClock(time.Second, func() time.Time { return now })
	th.Allow()
	if !th.LastScan().Equal(now) {
		t.Fatalf("expected LastScan=%v, got %v", now, th.LastScan())
	}
}

func TestThrottleLastScanNotUpdatedWhenBlocked(t *testing.T) {
	base := time.Now()
	calls := 0
	clock := func() time.Time {
		calls++
		if calls == 1 {
			return base
		}
		return base.Add(1 * time.Second) // still within the 5s gap
	}
	th := newThrottleWithClock(5*time.Second, clock)

	th.Allow() // sets LastScan to base
	th.Allow() // blocked — LastScan should remain base

	if !th.LastScan().Equal(base) {
		t.Fatalf("expected LastScan to remain %v when blocked, got %v", base, th.LastScan())
	}
}
