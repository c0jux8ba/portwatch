package notify_test

import (
	"testing"

	"github.com/user/portwatch/internal/notify"
	"github.com/user/portwatch/internal/ports"
)

func TestThresholdSkipsEmptyDiff(t *testing.T) {
	called := false
	inner := &capturingNotifier{fn: func(ports.Diff) error { called = true; return nil }}
	n := notify.NewThresholdNotifier(inner, 1)
	_ = n.Notify(ports.Diff{})
	if called {
		t.Fatal("should not notify on empty diff")
	}
}

func TestThresholdBlocksBelowMinimum(t *testing.T) {
	called := false
	inner := &capturingNotifier{fn: func(ports.Diff) error { called = true; return nil }}
	n := notify.NewThresholdNotifier(inner, 3)
	_ = n.Notify(ports.Diff{Opened: []int{8080, 9090}})
	if called {
		t.Fatal("should not notify when changes < threshold")
	}
}

func TestThresholdForwardsAtMinimum(t *testing.T) {
	called := false
	inner := &capturingNotifier{fn: func(ports.Diff) error { called = true; return nil }}
	n := notify.NewThresholdNotifier(inner, 2)
	_ = n.Notify(ports.Diff{Opened: []int{8080, 9090}})
	if !called {
		t.Fatal("should notify when changes == threshold")
	}
}

func TestThresholdForwardsAboveMinimum(t *testing.T) {
	called := false
	inner := &capturingNotifier{fn: func(ports.Diff) error { called = true; return nil }}
	n := notify.NewThresholdNotifier(inner, 2)
	_ = n.Notify(ports.Diff{Opened: []int{80, 443, 8080}})
	if !called {
		t.Fatal("should notify when changes > threshold")
	}
}

func TestThresholdCountsOpenedAndClosed(t *testing.T) {
	called := false
	inner := &capturingNotifier{fn: func(ports.Diff) error { called = true; return nil }}
	n := notify.NewThresholdNotifier(inner, 3)
	_ = n.Notify(ports.Diff{Opened: []int{8080}, Closed: []int{22, 443}})
	if !called {
		t.Fatal("should count both opened and closed ports toward threshold")
	}
}

func TestThresholdZeroTreatedAsOne(t *testing.T) {
	called := false
	inner := &capturingNotifier{fn: func(ports.Diff) error { called = true; return nil }}
	n := notify.NewThresholdNotifier(inner, 0)
	_ = n.Notify(ports.Diff{Opened: []int{80}})
	if !called {
		t.Fatal("threshold of 0 should be treated as 1")
	}
}
