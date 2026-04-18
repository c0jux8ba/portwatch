package notify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func makeOpsGenieNotifier(url string) *OpsGenieNotifier {
	return NewOpsGenieNotifier(url, "test-api-key", "portwatch-test")
}

func TestOpsGenieSkipsEmptyDiff(t *testing.T) {
	n := makeOpsGenieNotifier("http://localhost")
	if err := n.Notify(ports.Diff{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOpsGenieSuccess(t *testing.T) {
	var received map[string]interface{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	n := makeOpsGenieNotifier(ts.URL)
	d := ports.Diff{Opened: []int{8080}, Closed: []int{}}
	if err := n.Notify(d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received["alias"] == nil {
		t.Error("expected alias field in payload")
	}
}

func TestOpsGenieNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	n := makeOpsGenieNotifier(ts.URL)
	d := ports.Diff{Opened: []int{9090}}
	if err := n.Notify(d); err == nil {
		t.Error("expected error on non-2xx response")
	}
}

func TestOpsGenieDescriptionContainsPorts(t *testing.T) {
	desc := buildOpsGenieDescription(ports.Diff{Opened: []int{22, 80}, Closed: []int{443}})
	for _, want := range []string{"22", "80", "443"} {
		if !contains(desc, want) {
			t.Errorf("description missing port %s", want)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsHelper(s, sub))
}

func containsHelper(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
