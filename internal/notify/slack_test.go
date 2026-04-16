package notify

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func TestSlackNotifySuccess(t *testing.T) {
	var received slackPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	f := &fixedFormatter{text: "port 8080 opened"}
	n := NewSlackNotifier(ts.URL, f)

	d := ports.Diff{Opened: []int{8080}}
	if err := n.Notify(d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Text != "port 8080 opened" {
		t.Errorf("expected formatted text, got %q", received.Text)
	}
}

func TestSlackNotifySkipsEmptyDiff(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := NewSlackNotifier(ts.URL, &fixedFormatter{text: "x"})
	if err := n.Notify(ports.Diff{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected no HTTP call for empty diff")
	}
}

func TestSlackNotifyNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := NewSlackNotifier(ts.URL, &fixedFormatter{text: "alert"})
	d := ports.Diff{Opened: []int{9090}}
	if err := n.Notify(d); err == nil {
		t.Error("expected error on non-2xx response")
	}
}

// fixedFormatter returns a constant string regardless of diff.
type fixedFormatter struct{ text string }

func (f *fixedFormatter) Format(_ ports.Diff) string { return f.text }
