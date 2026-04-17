package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/user/portwatch/internal/ports"
)

const pagerDutyEventURL = "https://events.pagerduty.com/v2/enqueue"

type pagerDutyNotifier struct {
	integrationKey string
	client         *http.Client
	formatter      *Formatter
}

type pdPayload struct {
	RoutingKey  string    `json:"routing_key"`
	EventAction string    `json:"event_action"`
	Payload     pdDetails `json:"payload"`
}

type pdDetails struct {
	Summary  string `json:"summary"`
	Source   string `json:"source"`
	Severity string `json:"severity"`
}

// NewPagerDutyNotifier creates a Notifier that sends events to PagerDuty.
func NewPagerDutyNotifier(integrationKey string) Notifier {
	return &pagerDutyNotifier{
		integrationKey: integrationKey,
		client:         &http.Client{},
		formatter:      NewFormatter(""),
	}
}

func (p *pagerDutyNotifier) Notify(diff ports.Diff) error {
	if diff.IsEmpty() {
		return nil
	}

	summary := p.formatter.Format(diff)

	body := pdPayload{
		RoutingKey:  p.integrationKey,
		EventAction: "trigger",
		Payload: pdDetails{
			Summary:  summary,
			Source:   "portwatch",
			Severity: "warning",
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("pagerduty: marshal: %w", err)
	}

	resp, err := p.client.Post(pagerDutyEventURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("pagerduty: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("pagerduty: unexpected status %d", resp.StatusCode)
	}
	return nil
}
