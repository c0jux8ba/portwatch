package notify

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

const defaultTmpl = `{"host":"{{.Hostname}}","opened":{{len .Opened}},"closed":{{len .Closed}}}`

func makeWTNotifier(t *testing.T, url, tmpl string) *WebhookTemplateNotifier {
	t.Helper()
	n, err := NewWebhookTemplateNotifier(url, tmpl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return n
}

func TestWebhookTemplateSkipsEmptyDiff(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer ts.Close()
	n := makeWTNotifier(t, ts.URL, defaultTmpl)
	_ = n.Notify(ports.Diff{})
	if called {
		t.Fatal("expected no request for empty diff")
	}
}

func TestWebhookTemplateSuccess(t *testing.T) {
	var body []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.WriteHeader(200)
	}))
	defer ts.Close()
	n := makeWTNotifier(t, ts.URL, defaultTmpl)
	n.hostname = "testhost"
	err := n.Notify(ports.Diff{Opened: []int{8080}, Closed: []int{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("body not valid JSON: %v", err)
	}
	if m["host"] != "testhost" {
		t.Errorf("expected host testhost, got %v", m["host"])
	}
}

func TestWebhookTemplateNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer ts.Close()
	n := makeWTNotifier(t, ts.URL, defaultTmpl)
	err := n.Notify(ports.Diff{Opened: []int{9090}})
	if err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}

func TestWebhookTemplateInvalidTemplate(t *testing.T) {
	_, err := NewWebhookTemplateNotifier("http://localhost", `{{.Unclosed`)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestWebhookTemplateInvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer ts.Close()
	// Template renders non-JSON
	n := makeWTNotifier(t, ts.URL, `plain text {{.Hostname}}`)
	err := n.Notify(ports.Diff{Opened: []int{1234}})
	if err == nil {
		t.Fatal("expected error for invalid JSON output")
	}
}
