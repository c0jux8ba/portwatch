package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/user/portwatch/internal/ports"
)

// WebhookPayload is the JSON body sent to the webhook endpoint.
type WebhookPayload struct {
	Timestamp string       `json:"timestamp"`
	Opened    []ports.Port `json:"opened"`
	Closed    []ports.Port `json:"closed"`
}

// WebhookNotifier sends port change alerts to an HTTP endpoint.
type WebhookNotifier struct {
	URL    string
	Client *http.Client
}

// NewWebhookNotifier creates a WebhookNotifier with a sensible timeout.
func NewWebhookNotifier(url string) *WebhookNotifier {
	return &WebhookNotifier{
		URL: url,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Notify sends a webhook POST request describing opened and closed ports.
// It returns an error if the request fails or the server responds with a
// non-2xx status code.
func (w *WebhookNotifier) Notify(diff ports.Diff) error {
	if len(diff.Opened) == 0 && len(diff.Closed) == 0 {
		return nil
	}

	payload := WebhookPayload{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Opened:    diff.Opened,
		Closed:    diff.Closed,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshal payload: %w", err)
	}

	resp, err := w.Client.Post(w.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: post request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status %d from %s", resp.StatusCode, w.URL)
	}

	return nil
}
