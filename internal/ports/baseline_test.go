package ports

import (
	"os"
	"path/filepath"
	"testing"
)

func tempBaselinePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "baseline.json")
}

func TestBaselineMissingFileReturnsNil(t *testing.T) {
	b := NewBaseline(tempBaselinePath(t))
	entry, err := b.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry != nil {
		t.Fatalf("expected nil entry for missing file, got %+v", entry)
	}
}

func TestBaselineExistsAfterSave(t *testing.T) {
	b := NewBaseline(tempBaselinePath(t))
	if b.Exists() {
		t.Fatal("expected Exists() == false before Save")
	}
	if err := b.Save([]int{80, 443}); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if !b.Exists() {
		t.Fatal("expected Exists() == true after Save")
	}
}

func TestBaselineSaveAndLoad(t *testing.T) {
	b := NewBaseline(tempBaselinePath(t))
	want := []int{22, 80, 8080}
	if err := b.Save(want); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	entry, err := b.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if entry == nil {
		t.Fatal("expected non-nil entry")
	}
	if len(entry.Ports) != len(want) {
		t.Fatalf("port count mismatch: got %d want %d", len(entry.Ports), len(want))
	}
	for i, p := range want {
		if entry.Ports[i] != p {
			t.Errorf("port[%d]: got %d want %d", i, entry.Ports[i], p)
		}
	}
	if entry.RecordedAt.IsZero() {
		t.Error("expected non-zero RecordedAt")
	}
}

func TestBaselineOverwrite(t *testing.T) {
	path := tempBaselinePath(t)
	b := NewBaseline(path)

	_ = b.Save([]int{80})
	_ = b.Save([]int{443, 8443})

	entry, err := b.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(entry.Ports) != 2 {
		t.Fatalf("expected 2 ports after overwrite, got %d", len(entry.Ports))
	}
}

func TestBaselineCorruptFileReturnsError(t *testing.T) {
	path := tempBaselinePath(t)
	if err := os.WriteFile(path, []byte("not json{"), 0o644); err != nil {
		t.Fatal(err)
	}
	b := NewBaseline(path)
	_, err := b.Load()
	if err == nil {
		t.Fatal("expected error for corrupt baseline file")
	}
}
