package notify

import (
	"errors"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

type recordNotifier struct {
	calls []ports.Diff
	err   error
}

func (r *recordNotifier) Notify(d ports.Diff) error {
	r.calls = append(r.calls, d)
	return r.err
}

func TestBufferAccumulatesWithoutForwarding(t *testing.T) {
	rec := &recordNotifier{}
	b := NewBufferedNotifier(rec)

	_ = b.Notify(ports.Diff{Opened: []int{80}})
	_ = b.Notify(ports.Diff{Opened: []int{443}})

	if len(rec.calls) != 0 {
		t.Fatalf("expected 0 delegate calls before flush, got %d", len(rec.calls))
	}
	if b.Len() != 2 {
		t.Fatalf("expected Len 2, got %d", b.Len())
	}
}

func TestBufferFlushCombinesDiffs(t *testing.T) {
	rec := &recordNotifier{}
	b := NewBufferedNotifier(rec)

	_ = b.Notify(ports.Diff{Opened: []int{80}})
	_ = b.Notify(ports.Diff{Closed: []int{22}})

	if err := b.Flush(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rec.calls) != 1 {
		t.Fatalf("expected 1 delegate call, got %d", len(rec.calls))
	}
	if len(rec.calls[0].Opened) != 1 || rec.calls[0].Opened[0] != 80 {
		t.Errorf("unexpected opened: %v", rec.calls[0].Opened)
	}
	if len(rec.calls[0].Closed) != 1 || rec.calls[0].Closed[0] != 22 {
		t.Errorf("unexpected closed: %v", rec.calls[0].Closed)
	}
}

func TestBufferFlushResetsState(t *testing.T) {
	rec := &recordNotifier{}
	b := NewBufferedNotifier(rec)

	_ = b.Notify(ports.Diff{Opened: []int{8080}})
	_ = b.Flush()
	_ = b.Flush() // second flush should be no-op

	if len(rec.calls) != 1 {
		t.Fatalf("expected 1 call total, got %d", len(rec.calls))
	}
	if b.Len() != 0 {
		t.Fatalf("expected Len 0 after flush, got %d", b.Len())
	}
}

func TestBufferFlushPropagatesError(t *testing.T) {
	expected := errors.New("delegate error")
	rec := &recordNotifier{err: expected}
	b := NewBufferedNotifier(rec)

	_ = b.Notify(ports.Diff{Opened: []int{9090}})
	if err := b.Flush(); !errors.Is(err, expected) {
		t.Fatalf("expected delegate error, got %v", err)
	}
}

func TestBufferSkipsEmptyDiff(t *testing.T) {
	rec := &recordNotifier{}
	b := NewBufferedNotifier(rec)

	_ = b.Notify(ports.Diff{})
	if b.Len() != 0 {
		t.Fatalf("expected Len 0 for empty diff, got %d", b.Len())
	}
}
