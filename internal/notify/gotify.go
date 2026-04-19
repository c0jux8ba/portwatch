package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/user/portwatch/internal/ports"
)

// GotifyNotifier sends port change alerts to a Gotify server.
type GotifyNotifier struct {
	serverURL string
	token     string
	priority  int
	hostname  string
	client    *http.Client
}

type gotifyPayload struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	Priority int    `json:"priority"`
}

// NewGotifyNotifier creates a notifier that pushes to a Gotify server.
func NewGotifyNotifier(serverURL, token string, priority int) *GotifyNotifier {
	host, _ := os.Hostname()
	return &GotifyNotifier{
		serverURL: serverURL,
		token:     token,
		priority:  priority,
		hostname:  host,
		client:    &http.Client{},
	}
}

func (g *GotifyNotifier) Notify(diff ports.Diff) error {
	if diff.IsEmpty() {
		return nil
	}

	body := buildGotifyMessage(diff, g.hostname)
	payload := gotifyPayload{
		Title:    fmt.Sprintf("portwatch – %s", g.hostname),
		Message:  body,
		Priority: g.priority,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/message?token=%s", g.serverURL, g.token)
	resp, err := g.client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("gotify: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func buildGotifyMessage(diff ports.Diff, hostname string) string {
	var buf bytes.Buffer
	if len(diff.Opened) > 0 {
		fmt.Fprintf(&buf, "Opened: %v\n", diff.Opened)
	}
	if len(diff.Closed) > 0 {
		fmt.Fprintf(&buf, "Closed: %v\n", diff.Closed)
	}
	return buf.String()
}
