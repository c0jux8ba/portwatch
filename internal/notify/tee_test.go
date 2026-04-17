package notify

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func TestTeeSkipsEmptyDiff(t *testing.T) {
	var buf bytes.Buffer
	te := NewTeeNotifier(NewNoop(), &buf)
	_ = te.Notify(ports.Diff{})
	if buf.Len() != 0 {
		t.Fatalf("expected no output for empty diff, got %q", buf.String())
	}
}

func TestTeeWritesJSONRecord(t *testing.T) {
	var buf bytes.Buffer
	te := NewTeeNotifier(NewNoop(), &buf)
	d := ports.Diff{Opened: []int{8080}, Closed: []int{22}}
	if err := te.Notify(d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	line := strings.TrimSpace(buf.String())
	var rec map[string]interface{}
	if err := json.Unmarshal([]byte(line), &rec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if rec["timestamp"] == "" {
		t.Error("expected timestamp in record")
	}
	opened := rec["opened"].([]interface{})
	if len(opened) != 1 || opened[0].(float64) != 8080 {
		t.Errorf("unexpected opened: %v", opened)
	}
}

func TestTeeForwardsToWrapped(t *testing.T) {
	called := false
	n := notifierFunc(func(d ports.Diff) error { called = true; return nil })
	te := NewTeeNotifier(n, nil)
	_ = te.Notify(ports.Diff{Opened: []int{443}})
	if !called {
		t.Error("expected wrapped notifier to be called")
	}
}

func TestTeeReturnsWrappedError(t *testing.T) {
	want := errors.New("boom")
	n := notifierFunc(func(d ports.Diff) error { return want })
	te := NewTeeNotifier(n, nil)
	got := te.Notify(ports.Diff{Opened: []int{80}})
	if got != want {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestTeeNilWriterDoesNotPanic(t *testing.T) {
	te := NewTeeNotifier(NewNoop(), nil)
	if err := te.Notify(ports.Diff{Opened: []int{9000}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
