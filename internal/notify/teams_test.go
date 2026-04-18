package notify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func makeTeamsNotifier(url string) *TeamsNotifier {
	return NewTeamsNotifier(url, "test-host")
}

func TestTeamsNotifySkipsEmptyDiff(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer ts.Close()

	n := makeTeamsNotifier(ts.URL)
	if err := n.Notify(ports.Diff{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("expected no HTTP call for empty diff")
	}
}

func TestTeamsNotifySuccess(t *testing.T) {
	var got teamsPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := makeTeamsNotifier(ts.URL)
	diff := ports.Diff{Opened: []int{8080}, Closed: []int{22}}
	if err := n.Notify(diff); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Type != "MessageCard" {
		t.Errorf("expected MessageCard, got %q", got.Type)
	}
	if got.Title == "" {
		t.Error("expected non-empty title")
	}
}

func TestTeamsNotifyNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := makeTeamsNotifier(ts.URL)
	diff := ports.Diff{Opened: []int{9090}}
	if err := n.Notify(diff); err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}

func TestBuildTeamsBodyOpenedOnly(t *testing.T) {
	body := buildTeamsBody(ports.Diff{Opened: []int{80, 443}}, "host1")
	if body == "" {
		t.Fatal("expected non-empty body")
	}
	for _, want := range []string{"80", "443", "Opened"} {
		if !containsStr(body, want) {
			t.Errorf("body missing %q: %s", want, body)
		}
	}
}

func TestTeamsDefaultHostname(t *testing.T) {
	n := NewTeamsNotifier("http://example.com", "")
	if n.hostname == "" {
		t.Error("expected default hostname to be set")
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && stringContains(s, sub))
}

func stringContains(s, sub string) bool {
	for i := 0; i <= i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
