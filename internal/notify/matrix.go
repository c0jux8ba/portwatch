package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/user/portwatch/internal/ports"
)

// matrixMessage is the payload sent to the Matrix client-server API.
type matrixMessage struct {
	MsgType       string `json:"msgtype"`
	Body          string `json:"body"`
	FormattedBody string `json:"formatted_body,omitempty"`
	Format        string `json:"format,omitempty"`
}

// MatrixNotifier sends port-change alerts to a Matrix room via the
// client-server API (PUT /_matrix/client/v3/rooms/{roomID}/send/m.room.message).
type MatrixNotifier struct {
	homeserver string
	roomID     string
	token      string
	hostname   string
	client     *http.Client
}

// NewMatrixNotifier constructs a MatrixNotifier.
//
//   - homeserver is the base URL of the Matrix homeserver, e.g. "https://matrix.org".
//   - roomID is the fully-qualified room ID, e.g. "!abc123:matrix.org".
//   - token is a valid Matrix access token for a user that has joined the room.
func NewMatrixNotifier(homeserver, roomID, token string) *MatrixNotifier {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}
	return &MatrixNotifier{
		homeserver: homeserver,
		roomID:     roomID,
		token:      token,
		hostname:   hostname,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// Notify sends a Matrix message when diff contains at least one change.
// It returns nil when diff is empty (no-op) and propagates any HTTP or
// serialisation errors otherwise.
func (m *MatrixNotifier) Notify(diff ports.Diff) error {
	if diff.IsEmpty() {
		return nil
	}

	body := buildMatrixBody(diff, m.hostname)
	msg := matrixMessage{
		MsgType:       "m.text",
		Body:          body,
		Format:        "org.matrix.custom.html",
		FormattedBody: "<pre>" + body + "</pre>",
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("matrix: marshal payload: %w", err)
	}

	// Use a timestamp-based transaction ID to ensure idempotency per the spec.
	txnID := fmt.Sprintf("portwatch-%d", time.Now().UnixNano())
	url := fmt.Sprintf("%s/_matrix/client/v3/rooms/%s/send/m.room.message/%s",
		m.homeserver, m.roomID, txnID)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("matrix: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.token)

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("matrix: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("matrix: unexpected status %d", resp.StatusCode)
	}
	return nil
}

// buildMatrixBody formats the diff into a human-readable plain-text message.
func buildMatrixBody(diff ports.Diff, hostname string) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("[portwatch] Port change detected on %s\n", hostname))
	if len(diff.Opened) > 0 {
		buf.WriteString(fmt.Sprintf("  Opened: %s\n", joinInts(diff.Opened)))
	}
	if len(diff.Closed) > 0 {
		buf.WriteString(fmt.Sprintf("  Closed: %s\n", joinInts(diff.Closed)))
	}
	return buf.String()
}
