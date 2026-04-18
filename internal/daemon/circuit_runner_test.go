package daemon

import (
	"errors"
	"testing"
	"time"

	"github.com/user/portwatch/internal/notify"
	"github.com/user/portwatch/internal/ports"
)

type stubRunner struct{ err error }

func (s *stubRunner) Tick() error { return s.err }

type captureNotifier struct {
	calls int
	err   error
}

func (c *captureNotifier) Notify(_ ports.Diff) error {
	c.calls++
	return c.err
}

func samplePortDiff() ports.Diff {
	return ports.Diff{Opened: []int{9090}, Closed: []int{}}
}

func TestCircuitRunnerTickDelegates(t *testing.T) {
	inner := &stubRunner{}
	cn := &captureNotifier{}
	cr := NewCircuitRunner(inner, cn, 3, time.Minute)
	if err := cr.Tick(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCircuitRunnerTickPropagatesError(t *testing.T) {
	inner := &stubRunner{err: errors.New("scan failed")}
	cn := &captureNotifier{}
	cr := NewCircuitRunner(inner, cn, 3, time.Minute)
	if err := cr.Tick(); err == nil {
		t.Fatal("expected error from inner tick")
	}
}

func TestCircuitRunnerNotifyOpensAfterFailures(t *testing.T) {
	cn := &captureNotifier{err: errors.New("webhook down")}
	cr := NewCircuitRunner(&stubRunner{}, cn, 2, time.Minute)
	d := samplePortDiff()
	cr.NotifyDiff(d)
	cr.NotifyDiff(d)
	err := cr.NotifyDiff(d)
	if !errors.Is(err, notify.ErrCircuitOpen) {
		t.Fatalf("expected circuit open, got %v", err)
	}
}

func TestCircuitRunnerNotifySuccessPassesThrough(t *testing.T) {
	cn := &captureNotifier{}
	cr := NewCircuitRunner(&stubRunner{}, cn, 3, time.Minute)
	if err := cr.NotifyDiff(samplePortDiff()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cn.calls != 1 {
		t.Fatalf("expected 1 notify call, got %d", cn.calls)
	}
}
