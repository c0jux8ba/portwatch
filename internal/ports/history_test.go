package ports

import (
	"os"
	"path/filepath"
	"testing"
)

func tempHistoryPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "history.json")
}

func TestHistoryRecordAndReload(t *testing.T) {
	p := tempHistoryPath(t)
	h, err := NewHistory(p, 100)
	if err != nil {
		t.Fatalf("NewHistory: %v", err)
	}
	if err := h.Record([]int{8080}, nil); err != nil {
		t.Fatalf("Record: %v", err)
	}

	// Reload from disk
	h2, err := NewHistory(p, 100)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	events := h2.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if len(events[0].Opened) != 1 || events[0].Opened[0] != 8080 {
		t.Errorf("unexpected opened: %v", events[0].Opened)
	}
}

func TestHistorySkipsEmptyRecord(t *testing.T) {
	p := tempHistoryPath(t)
	h, _ := NewHistory(p, 100)
	_ = h.Record(nil, nil)
	if len(h.Events()) != 0 {
		t.Error("expected no events for empty record")
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Error("expected no file written for empty record")
	}
}

func TestHistoryMaxLen(t *testing.T) {
	p := tempHistoryPath(t)
	h, _ := NewHistory(p, 3)
	for i := 0; i < 5; i++ {
		_ = h.Record([]int{i + 1}, nil)
	}
	if len(h.Events()) != 3 {
		t.Errorf("expected 3 events, got %d", len(h.Events()))
	}
}

func TestHistoryMissingFileIsOK(t *testing.T) {
	p := tempHistoryPath(t)
	_, err := NewHistory(p, 50)
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
}

func TestHistoryEventsReturnsCopy(t *testing.T) {
	p := tempHistoryPath(t)
	h, _ := NewHistory(p, 10)
	_ = h.Record([]int{443}, []int{80})
	events := h.Events()
	events[0].Opened = nil // mutate copy
	if len(h.Events()[0].Opened) == 0 {
		t.Error("Events() should return a copy, not a reference")
	}
}

func TestHistoryRecordBothOpenedAndClosed(t *testing.T) {
	p := tempHistoryPath(t)
	h, _ := NewHistory(p, 10)
	if err := h.Record([]int{8080, 9090}, []int{80, 443}); err != nil {
		t.Fatalf("Record: %v", err)
	}
	events := h.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if len(events[0].Opened) != 2 {
		t.Errorf("expected 2 opened ports, got %d", len(events[0].Opened))
	}
	if len(events[0].Closed) != 2 {
		t.Errorf("expected 2 closed ports, got %d", len(events[0].Closed))
	}
}
