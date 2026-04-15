package ports

import (
	"fmt"
	"time"
)

// WatchEvent represents a change detected during a watch cycle.
type WatchEvent struct {
	Timestamp time.Time
	Diff      Diff
	ScanID    uint64
}

// Watcher continuously monitors ports and emits events on changes.
type Watcher struct {
	scanner  *Scanner
	filter   *Filter
	baseline *Baseline
	interval time.Duration
	events   chan WatchEvent
	stop     chan struct{}
	scanID   uint64
}

// NewWatcher creates a Watcher that scans the given port range at the
// specified interval, skipping ports excluded by filter.
func NewWatcher(scanner *Scanner, filter *Filter, baseline *Baseline, interval time.Duration) *Watcher {
	return &Watcher{
		scanner:  scanner,
		filter:   filter,
		baseline: baseline,
		interval: interval,
		events:   make(chan WatchEvent, 16),
		stop:     make(chan struct{}),
	}
}

// Events returns the read-only channel of detected change events.
func (w *Watcher) Events() <-chan WatchEvent {
	return w.events
}

// Start begins the watch loop in a background goroutine.
func (w *Watcher) Start() {
	go w.loop()
}

// Stop signals the watch loop to exit.
func (w *Watcher) Stop() {
	close(w.stop)
}

func (w *Watcher) loop() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-w.stop:
			return
		case <-ticker.C:
			if err := w.tick(); err != nil {
				_ = fmt.Errorf("watcher tick: %w", err)
			}
		}
	}
}

func (w *Watcher) tick() error {
	current, err := w.scanner.Scan()
	if err != nil {
		return err
	}
	if w.filter != nil {
		current = w.filter.Apply(current)
	}
	prior := w.baseline.Get()
	if prior == nil {
		w.baseline.Set(current)
		return nil
	}
	diff := Compare(prior, current)
	if diff.IsEmpty() {
		return nil
	}
	w.baseline.Set(current)
	w.scanID++
	w.events <- WatchEvent{
		Timestamp: time.Now(),
		Diff:      diff,
		ScanID:    w.scanID,
	}
	return nil
}
