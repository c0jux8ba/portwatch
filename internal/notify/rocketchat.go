package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/user/portwatch/internal/ports"
)

// rocketChatPayload is the incoming webhook payload format for Rocket.Chat.
type rocketChatPayload struct {
	Text        string             `json:"text"`
	Username    string             `json:"username,omitempty"`
	IconEmoji   string             `json:"emoji,omitempty"`
	Attachments []rocketAttachment `json:"attachments,omitempty"`
}

type rocketAttachment struct {
	Title  string `json:"title"`
	Text   string `json:"text"`
	Color  string `json:"color"`
}

// RocketChatNotifier sends port-change alerts to a Rocket.Chat channel via
// an incoming webhook integration.
type RocketChatNotifier struct {
	webhookURL string
	hostname   string
	username   string
	client     *http.Client
}

// NewRocketChatNotifier creates a RocketChatNotifier.
//
//   - webhookURL  – the Rocket.Chat incoming webhook URL (required)
//   - username    – display name shown in the channel; defaults to "portwatch"
//
// The caller's hostname is resolved once at construction time and embedded in
// every notification so alerts from multiple hosts are easy to distinguish.
func NewRocketChatNotifier(webhookURL, username string) *RocketChatNotifier {
	if username == "" {
		username = "portwatch"
	}
	host, _ := os.Hostname()
	return &RocketChatNotifier{
		webhookURL: webhookURL,
		hostname:   host,
		username:   username,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// Notify sends a Rocket.Chat message describing opened and closed ports.
// It is a no-op when diff carries no changes.
func (r *RocketChatNotifier) Notify(diff ports.Diff) error {
	if diff.IsEmpty() {
		return nil
	}

	payload := r.buildPayload(diff)
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("rocketchat: marshal payload: %w", err)
	}

	resp, err := r.client.Post(r.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("rocketchat: http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("rocketchat: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func (r *RocketChatNotifier) buildPayload(diff ports.Diff) rocketChatPayload {
	var attachments []rocketAttachment

	if len(diff.Opened) > 0 {
		attachments = append(attachments, rocketAttachment{
			Title: "🔓 Ports Opened",
			Text:  strings.Join(intsToStrings(diff.Opened), ", "),
			Color: "#e74c3c",
		})
	}
	if len(diff.Closed) > 0 {
		attachments = append(attachments, rocketAttachment{
			Title: "🔒 Ports Closed",
			Text:  strings.Join(intsToStrings(diff.Closed), ", "),
			Color: "#2ecc71",
		})
	}

	return rocketChatPayload{
		Text:        fmt.Sprintf("*portwatch* — port change detected on `%s`", r.hostname),
		Username:    r.username,
		IconEmoji:   ":satellite:",
		Attachments: attachments,
	}
}
