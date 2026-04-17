package notify

import "github.com/user/portwatch/internal/ports"

// ThresholdNotifier forwards a diff only when the total number of changed ports
// meets or exceeds a minimum threshold.
type ThresholdNotifier struct {
	inner     Notifier
	threshold int
}

// NewThresholdNotifier returns a ThresholdNotifier with the given minimum change count.
func NewThresholdNotifier(inner Notifier, threshold int) *ThresholdNotifier {
	if threshold < 1 {
		threshold = 1
	}
	return &ThresholdNotifier{inner: inner, threshold: threshold}
}

func (t *ThresholdNotifier) Notify(d ports.Diff) error {
	if d.IsEmpty() {
		return nil
	}
	if len(d.Opened)+len(d.Closed) < t.threshold {
		return nil
	}
	return t.inner.Notify(d)
}
