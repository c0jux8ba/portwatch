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

// DatadogNotifier sends port change events to Datadog as log events.
type DatadogNotifier struct {
	apiKey   string
	endpoint string
	host     string
	client   *http.Client
}

type datadogEvent struct {
	DdSource  string `json:"ddsource"`
	DdTags    string `json:"ddtags"`
	Hostname  string `json:"hostname"`
	Message   string `json:"message"`
	Service   string `json:"service"`
	Opened    []int  `json:"opened_ports,omitempty"`
	Closed    []int  `json:"closed_ports,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

func NewDatadogNotifier(apiKey, endpoint string) *DatadogNotifier {
	host, _ := os.Hostname()
	if endpoint == "" {
		endpoint = "https://http-intake.logs.datadoghq.com/api/v2/logs"
	}
	return &DatadogNotifier{
		apiKey:   apiKey,
		endpoint: endpoint,
		host:     host,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (d *DatadogNotifier) Notify(diff ports.Diff) error {
	if diff.IsEmpty() {
		return nil
	}
	msg := fmt.Sprintf("portwatch: %d opened, %d closed", len(diff.Opened), len(diff.Closed))
	evt := datadogEvent{
		DdSource:  "portwatch",
		DdTags:    "service:portwatch",
		Hostname:  d.host,
		Message:   msg,
		Service:   "portwatch",
		Opened:    diff.Opened,
		Closed:    diff.Closed,
		Timestamp: time.Now().Unix(),
	}
	body, err := json.Marshal([]datadogEvent{evt})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, d.endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("DD-API-KEY", d.apiKey)
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("datadog: unexpected status %d", resp.StatusCode)
	}
	return nil
}
