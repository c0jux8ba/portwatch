package daemon

import (
	"log"

	"github.com/user/portwatch/internal/ports"
)

// BaselineManager wraps ports.Baseline to integrate startup baseline
// logic into the daemon: on first run it saves a baseline and skips
// alerting; on subsequent runs it compares against the saved baseline.
type BaselineManager struct {
	baseline *ports.Baseline
	loaded   bool
	initial  []int
}

// NewBaselineManager creates a BaselineManager using the given file path.
func NewBaselineManager(path string) *BaselineManager {
	return &BaselineManager{
		baseline: ports.NewBaseline(path),
	}
}

// Init loads any existing baseline from disk. Call once at daemon startup.
func (m *BaselineManager) Init() error {
	entry, err := m.baseline.Load()
	if err != nil {
		return err
	}
	if entry != nil {
		m.initial = entry.Ports
		m.loaded = true
		log.Printf("[baseline] loaded %d ports from disk", len(m.initial))
	} else {
		log.Println("[baseline] no baseline on disk — will record on first scan")
	}
	return nil
}

// RecordIfNew saves the given ports as the baseline if none exists yet.
// Returns true if this was the first (baseline) scan.
func (m *BaselineManager) RecordIfNew(current []int) (isFirst bool, err error) {
	if m.loaded {
		return false, nil
	}
	if err := m.baseline.Save(current); err != nil {
		return true, err
	}
	m.initial = current
	m.loaded = true
	log.Printf("[baseline] recorded baseline with %d ports", len(current))
	return true, nil
}

// Initial returns the baseline port list (may be nil before first scan).
func (m *BaselineManager) Initial() []int {
	return m.initial
}

// IsLoaded reports whether a baseline has been established.
func (m *BaselineManager) IsLoaded() bool {
	return m.loaded
}
