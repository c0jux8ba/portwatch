package notify

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func makeSplunkNotifier(url string) *SplunkNotifier {
	n := NewSplunkNotifier(url, "test-token")
	n.host = "test-host"
	return n
}

func TestSplunkSkipsEmptyDiff(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer ts.Close()

	n := makeSplunkNotifier(ts.URL)
	if err := n.Notify(ports.Diff{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("expected no HTTP call for empty diff")
	}
}

func TestSplunkNotifySuccess(t *testing.T) {
	var received splunkEvent
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		if r.Header.Get("Authorization") != "Splunk test-token" {
			t.Errorf("missing or wrong auth header")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := makeSplunkNotifier(ts.URL)
	diff := ports.Diff{Opened: []int{8080}, Closed: []int{}}
	if err := n.Notify(diff); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Host != "test-host" {
		t.Errorf("expected host test-host, got %s", received.Host)
	}
	if received.Source != "portwatch" {
		t.Errorf("expected source portwatch, got %s", received.Source)
	}
}

func TestSplunkNotifyNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	n := makeSplunkNotifier(ts.URL)
	diff := ports.Diff{Opened: []int{443}}
	if err := n.Notify(diff); err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}

func TestSplunkEventContainsPorts(t *testing.T) {
	var received splunkEvent
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := makeSplunkNotifier(ts.URL)
	diff := ports.Diff{Opened: []int{22, 80}, Closed: []int{9090}}
	_ = n.Notify(diff)

	if received.Time == 0 {
		t.Error("expected non-zero time in event")
	}
	opened, ok := received.Event["opened"]
	if !ok || opened == nil {
		t.Error("expected opened field in event")
	}
}
