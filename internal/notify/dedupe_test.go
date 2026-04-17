package notify

import (
	"errors"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

type countingNotifier struct {
	calls int
	err   error
}

func (c *countingNotifier) Notify(_ ports.Diff) error {
	c.calls++
	return c.err
}

func TestDedupeSkipsEmptyDiff(t *testing.T) {
	inner := &countingNotifier{}
	d := NewDedupeNotifier(inner)
	_ = d.Notify(ports.Diff{})
	if inner.calls != 0 {
		t.Fatalf("expected 0 calls, got %d", inner.calls)
	}
}

func TestDedupeAllowsFirstNotification(t *testing.T) {
	inner := &countingNotifier{}
	d := NewDedupeNotifier(inner)
	diff := sampleDiff()
	_ = d.Notify(diff)
	if inner.calls != 1 {
		t.Fatalf("expected 1 call, got %d", inner.calls)
	}
}

func TestDedupeSuppressesIdenticalConsecutiveDiff(t *testing.T) {
	inner := &countingNotifier{}
	d := NewDedupeNotifier(inner)
	diff := sampleDiff()
	_ = d.Notify(diff)
	_ = d.Notify(diff)
	if inner.calls != 1 {
		t.Fatalf("expected 1 call, got %d", inner.calls)
	}
}

func TestDedupeAllowsDifferentDiff(t *testing.T) {
	inner := &countingNotifier{}
	d := NewDedupeNotifier(inner)
	_ = d.Notify(sampleDiff())
	_ = d.Notify(ports.Diff{Opened: []int{9090}, Closed: []int{}})
	if inner.calls != 2 {
		t.Fatalf("expected 2 calls, got %d", inner.calls)
	}
}

func TestDedupeAllowsAfterDifferentInterleaved(t *testing.T) {
	inner := &countingNotifier{}
	d := NewDedupeNotifier(inner)
	d1 := sampleDiff()
	d2 := ports.Diff{Opened: []int{9090}}
	_ = d.Notify(d1)
	_ = d.Notify(d2)
	_ = d.Notify(d1) // different from last, should go through
	if inner.calls != 3 {
		t.Fatalf("expected 3 calls, got %d", inner.calls)
	}
}

func TestDedupeForwardsError(t *testing.T) {
	expected := errors.New("send failed")
	inner := &countingNotifier{err: expected}
	d := NewDedupeNotifier(inner)
	err := d.Notify(sampleDiff())
	if !errors.Is(err, expected) {
		t.Fatalf("expected forwarded error, got %v", err)
	}
}
