package daemon

import (
	"fmt"
	"log"

	"github.com/user/portwatch/internal/notify"
	"github.com/user/portwatch/internal/ports"
)

// SyslogRunner wraps a Runner and attaches a SyslogNotifier so that every
// detected change is also written to syslog.
type SyslogRunner struct {
	inner  *Runner
	syslog *notify.SyslogNotifier
}

// NewSyslogRunner creates a SyslogRunner using the provided Runner and syslog tag.
func NewSyslogRunner(r *Runner, tag string) (*SyslogRunner, error) {
	sn, err := notify.NewSyslogNotifier(tag)
	if err != nil {
		return nil, fmt.Errorf("syslog runner: %w", err)
	}
	return &SyslogRunner{inner: r, syslog: sn}, nil
}

// Tick runs one scan cycle and additionally notifies syslog on changes.
func (sr *SyslogRunner) Tick() error {
	if err := sr.inner.Tick(); err != nil {
		return err
	}
	return nil
}

// NotifySyslog sends d to syslog directly; useful for injecting diffs in tests.
func (sr *SyslogRunner) NotifySyslog(d ports.Diff) error {
	if err := sr.syslog.Notify(d); err != nil {
		log.Printf("syslog notify error: %v", err)
		return err
	}
	return nil
}

// Close releases syslog resources.
func (sr *SyslogRunner) Close() error {
	return sr.syslog.Close()
}
