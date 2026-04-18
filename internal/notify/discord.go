package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/user/portwatch/internal/ports"
)

// DiscordNotifier sends port change alerts to a Discord webhook.
type DiscordNotifier struct {
	webhookURL string
	hostname   string
	client     *http.Client
}

type discordPayload struct {
	Content string `json:"content"`
}

// NewDiscordNotifier creates a DiscordNotifier targeting the given webhook URL.
func NewDiscordNotifier(webhookURL, hostname string) *DiscordNotifier {
	if hostname == "" {
		hostname = resolveHostname()
	}
	return &DiscordNotifier{
		webhookURL: webhookURL,
		hostname:   hostname,
		client:     &http.Client{},
	}
}

// Notify sends a Discord message if the diff is non-empty.
func (d *DiscordNotifier) Notify(diff ports.Diff) error {
	if diff.IsEmpty() {
		return nil
	}

	msg := fmt.Sprintf("**portwatch** [%s]\n", d.hostname)
	if len(diff.Opened) > 0 {
		msg += fmt.Sprintf("🟢 Opened: %s\n", joinInts(diff.Opened))
	}
	if len(diff.Closed) > 0 {
		msg += fmt.Sprintf("🔴 Closed: %s\n", joinInts(diff.Closed))
	}

	body, err := json.Marshal(discordPayload{Content: msg})
	if err != nil {
		return fmt.Errorf("discord: marshal: %w", err)
	}

	resp, err := d.client.Post(d.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("discord: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord: unexpected status %d", resp.StatusCode)
	}
	return nil
}
