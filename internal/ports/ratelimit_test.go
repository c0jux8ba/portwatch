package ports

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestRateLimiterAllowsFirstOccurrence(t *testing.T) {
	rl := NewRateLimiter(10 * time.Second)
	if !rl.Allow(8080) {
		t.Fatal("expected first occurrence to be allowed")
	}
}

func TestRateLimiterSuppressesWithinCooldown(t *testing.T) {
	base := time.Now()
	rl := NewRateLimiter(10 * time.Second)
	rl.now = fixedClock(base)
	rl.Allow(8080)

	// still within cooldown
	rl.now = fixedClock(base.Add(5 * time.Second))
	if rl.Allow(8080) {
		t.Fatal("expected port to be suppressed within cooldown")
	}
}

func TestRateLimiterAllowsAfterCooldown(t *testing.T) {
	base := time.Now()
	rl := NewRateLimiter(10 * time.Second)
	rl.now = fixedClock(base)
	rl.Allow(8080)

	// past cooldown
	rl.now = fixedClock(base.Add(11 * time.Second))
	if !rl.Allow(8080) {
		t.Fatal("expected port to be allowed after cooldown expires")
	}
}

func TestRateLimiterIndependentPorts(t *testing.T) {
	base := time.Now()
	rl := NewRateLimiter(10 * time.Second)
	rl.now = fixedClock(base)
	rl.Allow(8080)

	// different port should still be allowed
	if !rl.Allow(9090) {
		t.Fatal("expected different port to be allowed independently")
	}
}

func TestRateLimiterFilterDiff(t *testing.T) {
	base := time.Now()
	rl := NewRateLimiter(10 * time.Second)
	rl.now = fixedClock(base)

	// pre-seed port 80 so it is within cooldown
	rl.Allow(80)

	input := Diff{
		Opened: []int{80, 443},
		Closed: []int{8080},
	}
	out := rl.FilterDiff(input)

	if len(out.Opened) != 1 || out.Opened[0] != 443 {
		t.Fatalf("expected only port 443 in opened, got %v", out.Opened)
	}
	if len(out.Closed) != 1 || out.Closed[0] != 8080 {
		t.Fatalf("expected port 8080 in closed, got %v", out.Closed)
	}
}

func TestRateLimiterFilterDiffEmptyResult(t *testing.T) {
	base := time.Now()
	rl := NewRateLimiter(30 * time.Second)
	rl.now = fixedClock(base)
	rl.Allow(3000)

	input := Diff{Opened: []int{3000}, Closed: []int{}}
	out := rl.FilterDiff(input)

	if !out.IsEmpty() {
		t.Fatalf("expected empty diff after rate limiting, got %+v", out)
	}
}
