package notify

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func makeVictorOpsNotifier(baseURL string) *VictorOpsNotifier {
	return NewVictorOpsNotifier(baseURL, "team-portwatch", "portwatch-entity")
}

func TestVictorOpsSkipsEmptyDiff(t *testing.T) {
	n := makeVictorOpsNotifier("http://localhost")
	if err := n.Notify(ports.Diff{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVictorOpsSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := makeVictorOpsNotifier(ts.URL)
	if err := n.Notify(ports.Diff{Opened: []int{8080}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVictorOpsNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	n := makeVictorOpsNotifier(ts.URL)
	if err := n.Notify(ports.Diff{Opened: []int{22}}); err == nil {
		t.Error("expected error on non-2xx")
	}
}

func TestBuildVictorOpsMessageOpenedOnly(t *testing.T) {
	msg := buildVictorOpsMessage(ports.Diff{Opened: []int{80, 443}})
	if !containsHelper(msg, "Opened") {
		t.Error("expected 'Opened' in message")
	}
	if containsHelper(msg, "Closed") {
		t.Error("unexpected 'Closed' in message")
	}
}

func TestBuildVictorOpsMessageBothDirections(t *testing.T) {
	msg := buildVictorOpsMessage(ports.Diff{Opened: []int{9090}, Closed: []int{8080}})
	if !containsHelper(msg, "Opened") || !containsHelper(msg, "Closed") {
		t.Errorf("expected both directions in message, got: %s", msg)
	}
}
