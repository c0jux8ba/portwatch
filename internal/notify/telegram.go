package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/user/portwatch/internal/ports"
)

// TelegramNotifier sends port change alerts to a Telegram chat via Bot API.
type TelegramNotifier struct {
	token  string
	chatID string
	host   string
	client *http.Client
}

// NewTelegramNotifier creates a notifier that posts messages to the given Telegram chat.
func NewTelegramNotifier(token, chatID, host string) *TelegramNotifier {
	if host == "" {
		host, _ = resolveHostname()
	}
	return &TelegramNotifier{
		token:  token,
		chatID: chatID,
		host:   host,
		client: &http.Client{},
	}
}

func (t *TelegramNotifier) Notify(d ports.Diff) error {
	if d.IsEmpty() {
		return nil
	}

	text := buildTelegramText(t.host, d)

	body, err := json.Marshal(map[string]string{
		"chat_id":    t.chatID,
		"text":       text,
		"parse_mode": "Markdown",
	})
	if err != nil {
		return fmt.Errorf("telegram: marshal: %w", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.token)
	resp, err := t.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("telegram: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("telegram: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func buildTelegramText(host string, d ports.Diff) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("*portwatch* — `%s`\n", host))
	if len(d.Opened) > 0 {
		buf.WriteString(fmt.Sprintf("🟢 Opened: %s\n", joinInts(d.Opened)))
	}
	if len(d.Closed) > 0 {
		buf.WriteString(fmt.Sprintf("🔴 Closed: %s\n", joinInts(d.Closed)))
	}
	return buf.String()
}
