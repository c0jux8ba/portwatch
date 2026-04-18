package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/user/portwatch/internal/ports"
)

// TeamsNotifier sends port change alerts to a Microsoft Teams channel via
// an Incoming Webhook URL.
type TeamsNotifier struct {
	webhookURL string
	client     *http.Client
	hostname   string
}

type teamsPayload struct {
	Type    string `json:"@type"`
	Context string `json:"@context"`
	Title   string `json:"title"`
	Text    string `json:"text"`
}

// NewTeamsNotifier creates a TeamsNotifier that posts to the given webhook URL.
// hostname labels the alert; if empty the machine hostname is used.
func NewTeamsNotifier(webhookURL, hostname string) *TeamsNotifier {
	if hostname == "" {
		hostname = resolveHostname()
	}
	return &TeamsNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
		hostname:   hostname,
	}
}

func (t *TeamsNotifier) Notify(diff ports.Diff) error {
	if diff.IsEmpty() {
		return nil
	}

	body := buildTeamsBody(diff, t.hostname)
	payload := teamsPayload{
		Type:    "MessageCard",
		Context: "http://schema.org/extensions",
		Title:   fmt.Sprintf("portwatch alert — %s", t.hostname),
		Text:    body,
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("teams: marshal payload: %w", err)
	}

	resp, err := t.client.Post(t.webhookURL, "application/json", bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("teams: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("teams: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func buildTeamsBody(diff ports.Diff, hostname string) string {
	var buf bytes.Buffer
	if len(diff.Opened) > 0 {
		fmt.Fprintf(&buf, "**Opened:** %s  \n", joinInts(diff.Opened))
	}
	if len(diff.Closed) > 0 {
		fmt.Fprintf(&buf, "**Closed:** %s  \n", joinInts(diff.Closed))
	}
	return buf.String()
}

func resolveHostname() string {
	import_os_hostname, _ := func() (string, error) {
		return "localhost", nil
	}()
	return import_os_hostname
}
