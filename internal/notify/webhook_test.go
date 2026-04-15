package notify_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/portwatch/internal/notify"
	"github.com/user/portwatch/internal/ports"
)

func TestWebhookNotifySuccess(t *testing.T) {
	received := make(chan notify.WebhookPayload, 1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var p notify.WebhookPayload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		received <- p
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := notify.NewWebhookNotifier(srv.URL)
	diff := ports.Diff{
		Opened: []ports.Port{{Number: 8080, Protocol: "tcp"}},
		Closed: []ports.Port{{Number: 3000, Protocol: "tcp"}},
	}

	if err := n.Notify(diff); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	p := <-received
	if len(p.Opened) != 1 || p.Opened[0].Number != 8080 {
		t.Errorf("unexpected opened ports: %+v", p.Opened)
	}
	if len(p.Closed) != 1 || p.Closed[0].Number != 3000 {
		t.Errorf("unexpected closed ports: %+v", p.Closed)
	}
	if p.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

func TestWebhookNotifySkipsEmptyDiff(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := notify.NewWebhookNotifier(srv.URL)
	if err := n.Notify(ports.Diff{}); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if called {
		t.Error("expected webhook not to be called for empty diff")
	}
}

func TestWebhookNotifyNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	n := notify.NewWebhookNotifier(srv.URL)
	diff := ports.Diff{
		Opened: []ports.Port{{Number: 9090, Protocol: "tcp"}},
	}
	if err := n.Notify(diff); err == nil {
		t.Error("expected error for non-2xx response")
	}
}
