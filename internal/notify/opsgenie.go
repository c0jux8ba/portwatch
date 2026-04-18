package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/user/portwatch/internal/ports"
)

// OpsGenieNotifier sends alerts to OpsGenie when port changes are detected.
type OpsGenieNotifier struct {
	apiKey   string
	alias    string
	priority string
	hostname string
	client   *http.Client
}

type opsGeniePayload struct {
	Message     string            `json:"message"`
	Alias       string            `json:"alias"`
	Description string            `json:"description"`
	Priority    string            `json:"priority"`
	Details     map[string]string `json:"details"`
}

// NewOpsGenieNotifier creates a notifier that posts alerts to the OpsGenie
// Alerts API. priority defaults to "P3" if empty.
func NewOpsGenieNotifier(apiKey, alias, priority string) *OpsGenieNotifier {
	hostname, _ := os.Hostname()
	if priority == "" {
		priority = "P3"
	}
	if alias == "" {
		alias = "portwatch-alert"
	}
	return &OpsGenieNotifier{
		apiKey:   apiKey,
		alias:    alias,
		priority: priority,
		hostname: hostname,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

// Notify sends an OpsGenie alert if the diff contains changes.
func (o *OpsGenieNotifier) Notify(diff ports.Diff) error {
	if diff.IsEmpty() {
		return nil
	}

	message := fmt.Sprintf("portwatch: port change detected on %s", o.hostname)
	description := buildOpsGenieDescription(diff)

	payload := opsGeniePayload{
		Message:     message,
		Alias:       o.alias,
		Description: description,
		Priority:    o.priority,
		Details: map[string]string{
			"host":   o.hostname,
			"opened": fmt.Sprintf("%v", diff.Opened),
			"closed": fmt.Sprintf("%v", diff.Closed),
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("opsgenie: marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.opsgenie.com/v2/alerts", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("opsgenie: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "GenieKey "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("opsgenie: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("opsgenie: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func buildOpsGenieDescription(diff ports.Diff) string {
	var buf bytes.Buffer
	if len(diff.Opened) > 0 {
		fmt.Fprintf(&buf, "Opened ports: %v\n", diff.Opened)
	}
	if len(diff.Closed) > 0 {
		fmt.Fprintf(&buf, "Closed ports: %v\n", diff.Closed)
	}
	return buf.String()
}
