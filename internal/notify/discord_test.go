package notify

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func makeDiscordNotifier(url string) *DiscordNotifier {
	return NewDiscordNotifier(url, "testhost")
}

func TestDiscordNotifySkipsEmptyDiff(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer ts.Close()

	n := makeDiscordNotifier(ts.URL)
	if err := n.Notify(ports.Diff{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("expected no HTTP call for empty diff")
	}
}

func TestDiscordNotifySuccess(t *testing.T) {
	var got discordPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &got)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	n := makeDiscordNotifier(ts.URL)
	err := n.Notify(ports.Diff{Opened: []int{8080}, Closed: []int{22}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got.Content, "8080") {
		t.Errorf("expected content to contain 8080, got: %s", got.Content)
	}
	if !strings.Contains(got.Content, "22") {
		t.Errorf("expected content to contain 22, got: %s", got.Content)
	}
	if !strings.Contains(got.Content, "testhost") {
		t.Errorf("expected content to contain hostname, got: %s", got.Content)
	}
}

func TestDiscordNotifyNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := makeDiscordNotifier(ts.URL)
	err := n.Notify(ports.Diff{Opened: []int{9090}})
	if err == nil {
		t.Fatal("expected error for non-2xx response")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected error to mention status code, got: %v", err)
	}
}

func TestDiscordDefaultHostname(t *testing.T) {
	n := NewDiscordNotifier("http://example.com", "")
	if n.hostname == "" {
		t.Error("expected non-empty default hostname")
	}
}
