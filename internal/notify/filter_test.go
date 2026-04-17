package notify_test

import (
	"errors"
	"testing"

	"github.com/user/portwatch/internal/notify"
	"github.com/user/portwatch/internal/ports"
)

func TestFilterSkipsEmptyDiff(t *testing.T) {
	called := false
	n := notify.NewFilterNotifier(notify.NewNoop(), func(ports.Diff) bool {
		called = true
		return true
	})
	_ = n.Notify(ports.Diff{})
	if called {
		t.Fatal("predicate should not be called for empty diff")
	}
}

func TestFilterBlocksWhenPredicateFalse(t *testing.T) {
	var received *ports.Diff
	inner := &capturingNotifier{fn: func(d ports.Diff) error { received = &d; return nil }}
	n := notify.NewFilterNotifier(inner, func(ports.Diff) bool { return false })
	_ = n.Notify(ports.Diff{Opened: []int{9000}})
	if received != nil {
		t.Fatal("inner should not be called when predicate returns false")
	}
}

func TestFilterForwardsWhenPredicateTrue(t *testing.T) {
	var received *ports.Diff
	inner := &capturingNotifier{fn: func(d ports.Diff) error { received = &d; return nil }}
	n := notify.NewFilterNotifier(inner, func(ports.Diff) bool { return true })
	_ = n.Notify(ports.Diff{Opened: []int{8080}})
	if received == nil || len(received.Opened) != 1 {
		t.Fatal("inner should be called when predicate returns true")
	}
}

func TestFilterNilPredicateAlwaysForwards(t *testing.T) {
	var received *ports.Diff
	inner := &capturingNotifier{fn: func(d ports.Diff) error { received = &d; return nil }}
	n := notify.NewFilterNotifier(inner, nil)
	_ = n.Notify(ports.Diff{Opened: []int{443}})
	if received == nil {
		t.Fatal("nil predicate should forward all non-empty diffs")
	}
}

func TestFilterPropagatesError(t *testing.T) {
	want := errors.New("boom")
	inner := &capturingNotifier{fn: func(ports.Diff) error { return want }}
	n := notify.NewFilterNotifier(inner, func(ports.Diff) bool { return true })
	got := n.Notify(ports.Diff{Opened: []int{22}})
	if got != want {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

type capturingNotifier struct {
	fn func(ports.Diff) error
}

func (c *capturingNotifier) Notify(d ports.Diff) error { return c.fn(d) }
