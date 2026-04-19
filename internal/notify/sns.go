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

// SNSNotifier publishes port-change alerts to an AWS SNS topic via the
// SNS HTTP publish endpoint (useful when an AWS SDK is not desired).
type SNSNotifier struct {
	topicARN string
	endpoint  string
	subject   string
	client    *http.Client
	hostname  string
}

type snsPayload struct {
	TopicARN string `json:"TopicArn"`
	Subject  string `json:"Subject"`
	Message  string `json:"Message"`
}

// NewSNSNotifier creates an SNSNotifier. endpoint is the SNS HTTP URL,
// topicARN is the target topic, and subject is the notification subject.
func NewSNSNotifier(endpoint, topicARN, subject string) *SNSNotifier {
	h, _ := os.Hostname()
	if h == "" {
		h = "localhost"
	}
	return &SNSNotifier{
		topicARN: topicARN,
		endpoint:  endpoint,
		subject:   subject,
		client:    &http.Client{Timeout: 10 * time.Second},
		hostname:  h,
	}
}

func (s *SNSNotifier) Notify(d ports.Diff) error {
	if d.IsEmpty() {
		return nil
	}

	msg := fmt.Sprintf("host=%s opened=%v closed=%v", s.hostname, d.Opened, d.Closed)

	body, err := json.Marshal(snsPayload{
		TopicARN: s.topicARN,
		Subject:  s.subject,
		Message:  msg,
	})
	if err != nil {
		return fmt.Errorf("sns: marshal: %w", err)
	}

	resp, err := s.client.Post(s.endpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("sns: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("sns: unexpected status %d", resp.StatusCode)
	}
	return nil
}
