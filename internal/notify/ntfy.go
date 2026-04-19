package notify

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/user/portwatch/internal/ports"
)

// NtfyNotifier sends notifications to an ntfy.sh topic.
type NtfyNotifier struct {
	serverURL string
	topic     string
	hostname  string
	client    *http.Client
}

// NewNtfyNotifier creates a notifier that publishes to ntfy.sh (or a self-hosted instance).
// serverURL should be e.g. "https://ntfy.sh" and topic is the target topic name.
func NewNtfyNotifier(serverURL, topic string) *NtfyNotifier {
	host, _ := os.Hostname()
	if host == "" {
		host = "localhost"
	}
	return &NtfyNotifier{
		serverURL: strings.TrimRight(serverURL, "/"),
		topic:     topic,
		hostname:  host,
		client:    &http.Client{},
	}
}

func (n *NtfyNotifier) Notify(diff ports.Diff) error {
	if diff.IsEmpty() {
		return nil
	}

	body := buildNtfyMessage(diff, n.hostname)
	url := fmt.Sprintf("%s/%s", n.serverURL, n.topic)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString(body))
	if err != nil {
		return fmt.Errorf("ntfy: build request: %w", err)
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Title", fmt.Sprintf("portwatch alert on %s", n.hostname))
	req.Header.Set("Priority", "high")
	req.Header.Set("Tags", "warning,computer")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("ntfy: send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ntfy: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func buildNtfyMessage(diff ports.Diff, hostname string) string {
	var sb strings.Builder
	if len(diff.Opened) > 0 {
		fmt.Fprintf(&sb, "Opened ports on %s: %s\n", hostname, joinInts(diff.Opened))
	}
	if len(diff.Closed) > 0 {
		fmt.Fprintf(&sb, "Closed ports on %s: %s\n", hostname, joinInts(diff.Closed))
	}
	return strings.TrimSpace(sb.String())
}
