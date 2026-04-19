package daemon

import (
	"log"
	"time"

	"github.com/user/portwatch/internal/notify"
	"github.com/user/portwatch/internal/ports"
)

// BufferedRunner wraps a Watcher and flushes buffered notifications
// on a separate, longer interval than the scan interval.
type BufferedRunner struct {
	watcher  *ports.Watcher
	buffer   *notify.BufferedNotifier
	flushEvery time.Duration
	stopCh   chan struct{}
}

// NewBufferedRunner creates a runner that scans via watcher and flushes
// accumulated diffs to buffer every flushEvery duration.
func NewBufferedRunner(w *ports.Watcher, b *notify.BufferedNotifier, flushEvery time.Duration) *BufferedRunner {
	return &BufferedRunner{
		watcher:    w,
		buffer:     b,
		flushEvery: flushEvery,
		stopCh:     make(chan struct{}),
	}
}

// Run starts the flush loop; call Stop to terminate.
func (r *BufferedRunner) Run() {
	ticker := time.NewTicker(r.flushEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := r.buffer.Flush(); err != nil {
				log.Printf("buffered_runner: flush error: %v", err)
			}
		case <-r.stopCh:
			return
		}
	}
}

// Stop terminates the flush loop.
func (r *BufferedRunner) Stop() {
	close(r.stopCh)
}

// StopAndFlush stops the flush loop and performs one final flush to ensure
// no buffered notifications are lost on shutdown.
func (r *BufferedRunner) StopAndFlush() {
	r.Stop()
	if err := r.buffer.Flush(); err != nil {
		log.Printf("buffered_runner: final flush error: %v", err)
	}
}
