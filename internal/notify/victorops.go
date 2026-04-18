package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/user/portwatch/internal/ports"
)

// VictorOpsNotifier sends alerts to a VictorOps (Splunk On-Call) REST endpoint.
type VictorOpsNotifier struct {
	webhookURL string
	routingKey string
	entityID   string
	client     *http.Client
}

func NewVictorOpsNotifier(webhookURL, routingKey, entityID string) *VictorOpsNotifier {
	return &VictorOpsNotifier{
		webhookURL: webhookURL,
		routingKey: routingKey,
		entityID:   entityID,
		client:     &http.Client{},
	}
}

func (v *VictorOpsNotifier) Notify(d ports.Diff) error {
	if d.IsEmpty() {
		return nil
	}

	payload := map[string]string{
		"message_type":    "CRITICAL",
		"entity_id":       v.entityID,
		"entity_display_name": fmt.Sprintf("Port change detected [%s]", v.routingKey),
		"state_message":   buildVictorOpsMessage(d),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("victorops: marshal: %w", err)
	}

	url := strings.TrimRight(v.webhookURL, "/") + "/" + v.routingKey
	resp, err := v.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("victorops: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("victorops: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func buildVictorOpsMessage(d ports.Diff) string {
	var parts []string
	if len(d.Opened) > 0 {
		parts = append(parts, fmt.Sprintf("Opened: %s", joinInts(d.Opened)))
	}
	if len(d.Closed) > 0 {
		parts = append(parts, fmt.Sprintf("Closed: %s", joinInts(d.Closed)))
	}
	return strings.Join(parts, " | ")
}
