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

// SplunkNotifier sends port change events to a Splunk HTTP Event Collector.
type SplunkNotifier struct {
	endpoint string
	token    string
	host     string
	client   *http.Client
}

type splunkEvent struct {
	Time   float64        `json:"time"`
	Host   string         `json:"host"`
	Source string         `json:"source"`
	Event  map[string]any `json:"event"`
}

// NewSplunkNotifier creates a notifier that posts to a Splunk HEC endpoint.
// endpoint should be the full HEC URL, e.g. https://splunk:8088/services/collector.
func NewSplunkNotifier(endpoint, token string) *SplunkNotifier {
	host, _ := os.Hostname()
	return &SplunkNotifier{
		endpoint: endpoint,
		token:    token,
		host:     host,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *SplunkNotifier) Notify(diff ports.Diff) error {
	if diff.IsEmpty() {
		return nil
	}

	evt := splunkEvent{
		Time:   float64(time.Now().UnixMilli()) / 1000.0,
		Host:   s.host,
		Source: "portwatch",
		Event: map[string]any{
			"opened": diff.Opened,
			"closed": diff.Closed,
		},
	}

	body, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("splunk: marshal: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, s.endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("splunk: request: %w", err)
	}
	req.Header.Set("Authorization", "Splunk "+s.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("splunk: send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("splunk: unexpected status %d", resp.StatusCode)
	}
	return nil
}
