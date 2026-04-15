package notify

import (
	"errors"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

type stubNotifier struct {
	called bool
	err    error
}

func (s *stubNotifier) Notify(_ ports.Diff) error {
	s.called = true
	return s.err
}

func sampleDiff() ports.Diff {
	return ports.Diff{Opened: []int{8080}, Closed: []int{}}
}

func TestMultiCallsAllNotifiers(t *testing.T) {
	a, b := &stubNotifier{}, &stubNotifier{}
	m := NewMulti(a, b)
	if err := m.Notify(sampleDiff()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !a.called || !b.called {
		t.Error("expected both notifiers to be called")
	}
}

func TestMultiReturnsLastError(t *testing.T) {
	sentinel := errors.New("boom")
	a := &stubNotifier{err: sentinel}
	b := &stubNotifier{}
	m := NewMulti(a, b)
	err := m.Notify(sampleDiff())
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
	if !b.called {
		t.Error("second notifier should still be called despite first error")
	}
}

func TestMultiSkipsNilNotifiers(t *testing.T) {
	a := &stubNotifier{}
	m := NewMulti(nil, a, nil)
	if err := m.Notify(sampleDiff()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !a.called {
		t.Error("non-nil notifier should be called")
	}
}

func TestNoopNotifier(t *testing.T) {
	n := NewNoop()
	if err := n.Notify(sampleDiff()); err != nil {
		t.Errorf("noop should never return error, got %v", err)
	}
}
