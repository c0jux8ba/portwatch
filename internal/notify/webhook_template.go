package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/template"

	"github.com/user/portwatch/internal/ports"
)

// WebhookTemplateNotifier sends a webhook with a custom JSON body rendered from a Go template.
type WebhookTemplateNotifier struct {
	url      string
	tmpl     *template.Template
	client   *http.Client
	hostname string
}

type templateData struct {
	Hostname string
	Opened   []int
	Closed   []int
}

// NewWebhookTemplateNotifier creates a notifier that POSTs a templated body to url.
// tmplStr is a Go text/template producing valid JSON.
func NewWebhookTemplateNotifier(url, tmplStr string) (*WebhookTemplateNotifier, error) {
	tmpl, err := template.New("webhook").Parse(tmplStr)
	if err != nil {
		return nil, fmt.Errorf("webhook_template: parse template: %w", err)
	}
	host, _ := os.Hostname()
	return &WebhookTemplateNotifier{
		url:      url,
		tmpl:     tmpl,
		client:   &http.Client{},
		hostname: host,
	}, nil
}

func (w *WebhookTemplateNotifier) Notify(diff ports.Diff) error {
	if diff.IsEmpty() {
		return nil
	}
	data := templateData{
		Hostname: w.hostname,
		Opened:   diff.Opened,
		Closed:   diff.Closed,
	}
	var buf bytes.Buffer
	if err := w.tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("webhook_template: render: %w", err)
	}
	// Validate the rendered output is JSON.
	if !json.Valid(buf.Bytes()) {
		return fmt.Errorf("webhook_template: rendered body is not valid JSON")
	}
	resp, err := w.client.Post(w.url, "application/json", &buf)
	if err != nil {
		return fmt.Errorf("webhook_template: post: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook_template: unexpected status %d", resp.StatusCode)
	}
	return nil
}
