package notify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func makeMatrixNotifier(url, roomID, token string) *MatrixNotifier {
	n, _ := NewMatrixNotifier(url, roomID, token, "portwatch")
	return n
}

func TestMatrixSkipsEmptyDiff(t *testing.T) {
	n := makeMatrixNotifier("http://localhost", "!room:server", "token")
	if err := n.Notify(ports.Diff{}); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestMatrixNotifySuccess(t *testing.T) {
	var received map[string]interface{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"event_id":"$abc"}`))
	}))
	defer ts.Close()

	n := makeMatrixNotifier(ts.URL, "!room:server", "mytoken")
	diff := ports.Diff{Opened: []int{8080}, Closed: []int{}}
	if err := n.Notify(diff); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received["msgtype"] != "m.text" {
		t.Errorf("expected msgtype m.text, got %v", received["msgtype"])
	}
	body, _ := received["body"].(string)
	if body == "" {
		t.Error("expected non-empty body")
	}
}

func TestMatrixNotifyNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	n := makeMatrixNotifier(ts.URL, "!room:server", "token")
	diff := ports.Diff{Opened: []int{443}}
	if err := n.Notify(diff); err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}

func TestMatrixAuthHeader(t *testing.T) {
	var authHeader string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"event_id":"$x"}`))
	}))
	defer ts.Close()

	n := makeMatrixNotifier(ts.URL, "!room:server", "secret")
	n.Notify(ports.Diff{Opened: []int{22}})
	if authHeader != "Bearer secret" {
		t.Errorf("expected 'Bearer secret', got %q", authHeader)
	}
}

func TestMatrixBodyContainsPorts(t *testing.T) {
	var received map[string]interface{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"event_id":"$y"}`))
	}))
	defer ts.Close()

	n := makeMatrixNotifier(ts.URL, "!room:server", "tok")
	n.Notify(ports.Diff{Opened: []int{9200}, Closed: []int{6379}})
	body, _ := received["body"].(string)
	for _, want := range []string{"9200", "6379"} {
		if !contains(body, want) {
			t.Errorf("body missing %q: %s", want, body)
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
