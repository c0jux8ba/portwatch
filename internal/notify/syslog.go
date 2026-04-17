package notify

import (
	"fmt"
	"log/syslog"

	"github.com/user/portwatch/internal/ports"
)

// SyslogNotifier sends port change alerts to the system syslog.
type SyslogNotifier struct {
	writer *syslog.Writer
	tag    string
}

// NewSyslogNotifier creates a SyslogNotifier. tag is the syslog program tag.
func NewSyslogNotifier(tag string) (*SyslogNotifier, error) {
	if tag == "" {
		tag = "portwatch"
	}
	w, err := syslog.New(syslog.LOG_NOTICE|syslog.LOG_DAEMON, tag)
	if err != nil {
		return nil, fmt.Errorf("syslog: open: %w", err)
	}
	return &SyslogNotifier{writer: w, tag: tag}, nil
}

// Notify writes a syslog entry when ports have changed.
func (s *SyslogNotifier) Notify(d ports.Diff) error {
	if d.IsEmpty() {
		return nil
	}
	msg := buildSyslogMessage(d)
	return s.writer.Notice(msg)
}

// Close releases the underlying syslog connection.
func (s *SyslogNotifier) Close() error {
	return s.writer.Close()
}

func buildSyslogMessage(d ports.Diff) string {
	var opened, closed string
	if len(d.Opened) > 0 {
		opened = fmt.Sprintf("opened=%v", d.Opened)
	}
	if len(d.Closed) > 0 {
		closed = fmt.Sprintf("closed=%v", d.Closed)
	}
	switch {
	case opened != "" && closed != "":
		return fmt.Sprintf("port change detected: %s %s", opened, closed)
	case opened != "":
		return fmt.Sprintf("port change detected: %s", opened)
	default:
		return fmt.Sprintf("port change detected: %s", closed)
	}
}
