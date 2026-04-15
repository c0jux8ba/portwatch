package daemon

import (
	"path/filepath"
	"testing"
)

func tempBaselineMgrPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "baseline.json")
}

func TestBaselineManagerNoFileOnDisk(t *testing.T) {
	m := NewBaselineManager(tempBaselineMgrPath(t))
	if err := m.Init(); err != nil {
		t.Fatalf("Init error: %v", err)
	}
	if m.IsLoaded() {
		t.Fatal("expected IsLoaded==false when no file exists")
	}
	if m.Initial() != nil {
		t.Fatal("expected nil Initial() before first scan")
	}
}

func TestBaselineManagerRecordIfNewFirstRun(t *testing.T) {
	m := NewBaselineManager(tempBaselineMgrPath(t))
	_ = m.Init()

	ports := []int{22, 80, 443}
	isFirst, err := m.RecordIfNew(ports)
	if err != nil {
		t.Fatalf("RecordIfNew error: %v", err)
	}
	if !isFirst {
		t.Fatal("expected isFirst==true on first call")
	}
	if !m.IsLoaded() {
		t.Fatal("expected IsLoaded==true after RecordIfNew")
	}
	if len(m.Initial()) != len(ports) {
		t.Fatalf("Initial port count: got %d want %d", len(m.Initial()), len(ports))
	}
}

func TestBaselineManagerRecordIfNewSubsequentCall(t *testing.T) {
	m := NewBaselineManager(tempBaselineMgrPath(t))
	_ = m.Init()
	_, _ = m.RecordIfNew([]int{80})

	// second call should be a no-op
	isFirst, err := m.RecordIfNew([]int{80, 443})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isFirst {
		t.Fatal("expected isFirst==false on second call")
	}
	// initial should still reflect the first scan
	if len(m.Initial()) != 1 {
		t.Fatalf("Initial should not change after second call, got %v", m.Initial())
	}
}

func TestBaselineManagerLoadsExistingBaseline(t *testing.T) {
	path := tempBaselineMgrPath(t)

	// pre-populate via a first manager instance
	m1 := NewBaselineManager(path)
	_ = m1.Init()
	_, _ = m1.RecordIfNew([]int{22, 8080})

	// second instance should load from disk
	m2 := NewBaselineManager(path)
	if err := m2.Init(); err != nil {
		t.Fatalf("Init error: %v", err)
	}
	if !m2.IsLoaded() {
		t.Fatal("expected IsLoaded==true when baseline exists on disk")
	}
	if len(m2.Initial()) != 2 {
		t.Fatalf("expected 2 ports from disk, got %d", len(m2.Initial()))
	}
}
