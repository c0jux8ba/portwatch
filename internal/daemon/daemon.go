package daemon

import (
	"context"
	"log"
	"time"

	"github.com/user/portwatch/internal/config"
	"github.com/user/portwatch/internal/notify"
	"github.com/user/portwatch/internal/ports"
)

// Daemon periodically scans ports and notifies on changes.
type Daemon struct {
	cfg      *config.Config
	scanner  *ports.Scanner
	notifier notify.Notifier
	previous []int
}

// New creates a new Daemon from the given config and notifier.
func New(cfg *config.Config, notifier notify.Notifier) *Daemon {
	return &Daemon{
		cfg:      cfg,
		scanner:  ports.NewScanner(cfg.StartPort, cfg.EndPort, cfg.ScanTimeoutMs),
		notifier: notifier,
	}
}

// Run starts the daemon loop, blocking until ctx is cancelled.
func (d *Daemon) Run(ctx context.Context) error {
	log.Printf("portwatch: starting daemon (ports %d-%d, interval %ds)",
		d.cfg.StartPort, d.cfg.EndPort, d.cfg.IntervalSecs)

	// Perform an initial scan to establish baseline without alerting.
	initial, err := d.scanner.Scan()
	if err != nil {
		return err
	}
	d.previous = initial
	log.Printf("portwatch: baseline established, %d open ports", len(initial))

	ticker := time.NewTicker(time.Duration(d.cfg.IntervalSecs) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("portwatch: shutting down")
			return nil
		case <-ticker.C:
			if err := d.tick(); err != nil {
				log.Printf("portwatch: scan error: %v", err)
			}
		}
	}
}

func (d *Daemon) tick() error {
	current, err := d.scanner.Scan()
	if err != nil {
		return err
	}

	diff := ports.Compare(d.previous, current)
	if len(diff.Opened) > 0 || len(diff.Closed) > 0 {
		log.Printf("portwatch: change detected — opened: %v, closed: %v", diff.Opened, diff.Closed)
		if notifyErr := d.notifier.Notify(diff); notifyErr != nil {
			log.Printf("portwatch: notify error: %v", notifyErr)
		}
	}

	d.previous = current
	return nil
}
