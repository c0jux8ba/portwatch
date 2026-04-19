package notify

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func makeDatadogNotifier(url string) *DatadogNotifier {
	n := NewDatadogNotifier("test-api-key", url)
	n.host = "testhost"
	return n
}

func TestDatadogSkipsEmptyDiff(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer ts.Close()
	n := makeDatadogNotifier(ts.URL)
	if err := n.Notify(ports.Diff{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("expected no HTTP call for empty diff")
	}
}

func TestDatadogNotifySuccess(t *testing.T) {
	var body []map[string]interface{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("DD-API-KEY") != "test-api-key" {
			t.Error("missing or wrong API key header")
		}
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &body)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()
	n := makeDatadogNotifier(ts.URL)
	diff := ports.Diff{Opened: []int{8080}, Closed: []int{22}}
	if err := n.Notify(diff); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(body) != 1 {
		t.Fatalf("expected 1 event, got %d", len(body))
	}
	if body[0]["hostname"] != "testhost" {
		t.Errorf("unexpected hostname: %v", body[0]["hostname"])
	}
	if body[0]["ddsource"] != "portwatch" {
		t.Errorf("unexpected ddsource: %v", body[0]["ddsource"])
	}
}

func TestDatadogNotifyNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()
	n := makeDatadogNotifier(ts.URL)
	diff := ports.Diff{Opened: []int{9090}}
	if err := n.Notify(diff); err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}

func TestDatadogEventContainsPorts(t *testing.T) {
	var body []map[string]interface{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &body)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()
	n := makeDatadogNotifier(ts.URL)
	diff := ports.Diff{Opened: []int{443, 80}}
	_ = n.Notify(diff)
	if len(body) == 0 {
		t.Fatal("no events received")
	}
	opened, ok := body[0]["opened_ports"]
	if !ok || opened == nil {
		t.Error("expected opened_ports in payload")
	}
}
