package notify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func makeGotifyNotifier(serverURL string) *GotifyNotifier {
	n := NewGotifyNotifier(serverURL, "testtoken", 5)
	n.hostname = "testhost"
	return n
}

func TestGotifySkipsEmptyDiff(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer ts.Close()

	n := makeGotifyNotifier(ts.URL)
	if err := n.Notify(ports.Diff{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("expected no HTTP call for empty diff")
	}
}

func TestGotifyNotifySuccess(t *testing.T) {
	var received gotifyPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode error: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := makeGotifyNotifier(ts.URL)
	diff := ports.Diff{Opened: []int{8080}, Closed: []int{22}}
	if err := n.Notify(diff); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Priority != 5 {
		t.Errorf("expected priority 5, got %d", received.Priority)
	}
	if received.Title == "" {
		t.Error("expected non-empty title")
	}
}

func TestGotifyNotifyNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	n := makeGotifyNotifier(ts.URL)
	diff := ports.Diff{Opened: []int{9090}}
	if err := n.Notify(diff); err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}

func TestBuildGotifyMessageContainsPorts(t *testing.T) {
	diff := ports.Diff{Opened: []int{443}, Closed: []int{80}}
	msg := buildGotifyMessage(diff, "myhost")
	for _, want := range []string{"443", "80"} {
		if !containsStr(msg, want) {
			t.Errorf("expected message to contain %q, got: %s", want, msg)
		}
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && stringContains(s, sub))
}

func stringContains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
