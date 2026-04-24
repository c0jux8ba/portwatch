package notify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func makeMattermostNotifier(url string) *MattermostNotifier {
	n := NewMattermostNotifier(url, "portwatch", "#alerts")
	n.hostname = "testhost"
	return n
}

func TestMattermostSkipsEmptyDiff(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer ts.Close()

	n := makeMattermostNotifier(ts.URL)
	if err := n.Notify(ports.Diff{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("expected no HTTP call for empty diff")
	}
}

func TestMattermostNotifySuccess(t *testing.T) {
	var received mattermostPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := makeMattermostNotifier(ts.URL)
	diff := ports.Diff{Opened: []int{8080}, Closed: []int{22}}
	if err := n.Notify(diff); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Username != "portwatch" {
		t.Errorf("username = %q, want %q", received.Username, "portwatch")
	}
	if received.Channel != "#alerts" {
		t.Errorf("channel = %q, want %q", received.Channel, "#alerts")
	}
	if !strings.Contains(received.Text, "testhost") {
		t.Errorf("text missing hostname, got: %s", received.Text)
	}
}

func TestMattermostNotifyNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := makeMattermostNotifier(ts.URL)
	diff := ports.Diff{Opened: []int{9090}}
	if err := n.Notify(diff); err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}

func TestBuildMattermostTextOpenedOnly(t *testing.T) {
	diff := ports.Diff{Opened: []int{443, 80}}
	text := buildMattermostText(diff, "myhost")
	if !strings.Contains(text, "Opened") {
		t.Errorf("expected 'Opened' in text, got: %s", text)
	}
	if strings.Contains(text, "Closed") {
		t.Errorf("unexpected 'Closed' in text, got: %s", text)
	}
}

func TestBuildMattermostTextClosedOnly(t *testing.T) {
	diff := ports.Diff{Closed: []int{22}}
	text := buildMattermostText(diff, "myhost")
	if strings.Contains(text, "Opened") {
		t.Errorf("unexpected 'Opened' in text, got: %s", text)
	}
	if !strings.Contains(text, "Closed") {
		t.Errorf("expected 'Closed' in text, got: %s", text)
	}
}

func TestMattermostDefaultHostname(t *testing.T) {
	n := NewMattermostNotifier("http://example.com", "", "")
	if n.hostname == "" {
		t.Error("expected non-empty default hostname")
	}
}
