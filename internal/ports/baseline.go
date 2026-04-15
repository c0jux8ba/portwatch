package ports

import (
	"encoding/json"
	"os"
	"sync"
)

// Baseline holds the last-known set of open ports and optionally persists
// it to disk so it survives process restarts.
type Baseline struct {
	mu   sync.RWMutex
	data []int
	path string
}

// NewBaseline creates a Baseline backed by the given file path.
// Pass an empty string for an in-memory-only baseline.
func NewBaseline(path string) *Baseline {
	return &Baseline{path: path}
}

// Get returns the current baseline snapshot (nil if not yet set).
func (b *Baseline) Get() []int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.data == nil {
		return nil
	}
	out := make([]int, len(b.data))
	copy(out, b.data)
	return out
}

// Set updates the in-memory baseline and, if a path is configured,
// persists it to disk.
func (b *Baseline) Set(ports []int) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.data = make([]int, len(ports))
	copy(b.data, ports)
	if b.path == "" {
		return nil
	}
	return b.save()
}

// Load reads a previously persisted baseline from disk.
// Returns nil without error if the file does not exist.
func (b *Baseline) Load() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	raw, err := os.ReadFile(b.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, &b.data)
}

func (b *Baseline) save() error {
	raw, err := json.Marshal(b.data)
	if err != nil {
		return err
	}
	return os.WriteFile(b.path, raw, 0o644)
}
