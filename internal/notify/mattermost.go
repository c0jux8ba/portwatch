package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/user/portwatch/internal/ports"
)

// MattermostNotifier sends port-change alerts to a Mattermost incoming webhook.
type MattermostNotifier struct {
	webhookURL string
	username   string
	channel    string
	hostname   string
	client     *http.Client
}

type mattermostPayload struct {
	Username string `json:"username,omitempty"`
	Channel  string `json:"channel,omitempty"`
	Text     string `json:"text"`
}

// NewMattermostNotifier creates a notifier that posts to a Mattermost webhook.
// username and channel are optional (Mattermost webhook defaults apply when empty).
func NewMattermostNotifier(webhookURL, username, channel string) *MattermostNotifier {
	host, _ := os.Hostname()
	return &MattermostNotifier{
		webhookURL: webhookURL,
		username:   username,
		channel:    channel,
		hostname:   host,
		client:     &http.Client{},
	}
}

// Notify implements notify.Notifier.
func (m *MattermostNotifier) Notify(diff ports.Diff) error {
	if diff.IsEmpty() {
		return nil
	}

	text := buildMattermostText(diff, m.hostname)

	payload := mattermostPayload{
		Username: m.username,
		Channel:  m.channel,
		Text:     text,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("mattermost: marshal payload: %w", err)
	}

	resp, err := m.client.Post(m.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("mattermost: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("mattermost: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func buildMattermostText(diff ports.Diff, hostname string) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("**Port change detected on `%s`**\n", hostname))
	if len(diff.Opened) > 0 {
		buf.WriteString(fmt.Sprintf(":white_check_mark: **Opened:** %s\n", joinInts(diff.Opened)))
	}
	if len(diff.Closed) > 0 {
		buf.WriteString(fmt.Sprintf(":x: **Closed:** %s\n", joinInts(diff.Closed)))
	}
	return buf.String()
}
