package daemon

import (
	"fmt"

	"github.com/user/portwatch/internal/notify"
	"github.com/user/portwatch/internal/ports"
)

// WebhookTemplateRunner wires a WebhookTemplateNotifier into the daemon tick loop.
type WebhookTemplateRunner struct {
	scanner  *ports.Scanner
	baseline *BaselineManager
	notifier *notify.WebhookTemplateNotifier
}

// NewWebhookTemplateRunner creates a runner that alerts via a templated webhook.
func NewWebhookTemplateRunner(cfg RunnerConfig, url, tmplStr string) (*WebhookTemplateRunner, error) {
	n, err := notify.NewWebhookTemplateNotifier(url, tmplStr)
	if err != nil {
		return nil, fmt.Errorf("webhook_template_runner: %w", err)
	}
	return &WebhookTemplateRunner{
		scanner:  ports.NewScanner(cfg.StartPort, cfg.EndPort),
		baseline: NewBaselineManager(cfg.BaselinePath),
		notifier: n,
	}, nil
}

// Tick performs one scan cycle and fires the notifier on changes.
func (r *WebhookTemplateRunner) Tick() error {
	current, err := r.scanner.Scan()
	if err != nil {
		return fmt.Errorf("webhook_template_runner: scan: %w", err)
	}
	base := r.baseline.Get()
	if base == nil {
		r.baseline.Set(current)
		return nil
	}
	diff := ports.Compare(base, current)
	if diff.IsEmpty() {
		return nil
	}
	r.baseline.Set(current)
	return r.notifier.Notify(diff)
}
