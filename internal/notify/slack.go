package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/user/portwatch/internal/ports"
)

// SlackNotifier sends port change alerts to a Slack webhook URL.
type SlackNotifier struct {
	webhookURL string
	formatter  Formatter
	client     *http.Client
}

type slackPayload struct {
	Text string `json:"text"`
}

// NewSlackNotifier creates a SlackNotifier that posts to the given Slack webhook URL.
func NewSlackNotifier(webhookURL string, f Formatter) *SlackNotifier {
	return &SlackNotifier{
		webhookURL: webhookURL,
		formatter:  f,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// Notify sends a Slack message if the diff is non-empty.
func (s *SlackNotifier) Notify(d ports.Diff) error {
	if d.IsEmpty() {
		return nil
	}

	text := s.formatter.Format(d)
	payload := slackPayload{Text: text}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("slack: marshal payload: %w", err)
	}

	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("slack: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack: unexpected status %d", resp.StatusCode)
	}

	return nil
}
