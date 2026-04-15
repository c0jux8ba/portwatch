package ports

import (
	"encoding/json"
	"os"
	"time"
)

// BaselineEntry records a port snapshot taken at a specific time.
type BaselineEntry struct {
	Ports     []int     `json:"ports"`
	RecordedAt time.Time `json:"recorded_at"`
}

// Baseline persists and loads a reference port snapshot used to
// suppress alerts on the very first scan (daemon startup).
type Baseline struct {
	path string
}

// NewBaseline creates a Baseline backed by the given file path.
func NewBaseline(path string) *Baseline {
	return &Baseline{path: path}
}

// Save writes the given port list as the current baseline.
func (b *Baseline) Save(ports []int) error {
	entry := BaselineEntry{
		Ports:      ports,
		RecordedAt: time.Now().UTC(),
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(b.path, data, 0o644)
}

// Load reads the persisted baseline. Returns nil, nil if the file does
// not exist yet (first run).
func (b *Baseline) Load() (*BaselineEntry, error) {
	data, err := os.ReadFile(b.path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var entry BaselineEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

// Exists reports whether a baseline file has been saved.
func (b *Baseline) Exists() bool {
	_, err := os.Stat(b.path)
	return err == nil
}
