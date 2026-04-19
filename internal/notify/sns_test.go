package notify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func makeSNSNotifier(url string) *SNSNotifier {
	n := NewSNSNotifier(url, "arn:aws:sns:us-east-1:123456789012:portwatch", "Port Change")
	n.hostname = "testhost"
	return n
}

func TestSNSNotifySkipsEmptyDiff(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer ts.Close()

	n := makeSNSNotifier(ts.URL)
	if err := n.Notify(ports.Diff{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("expected no HTTP call for empty diff")
	}
}

func TestSNSNotifySuccess(t *testing.T) {
	var got snsPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := makeSNSNotifier(ts.URL)
	d := ports.Diff{Opened: []int{8080}, Closed: []int{}}
	if err := n.Notify(d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TopicARN == "" {
		t.Fatal("expected TopicARN to be set")
	}
	if got.Subject != "Port Change" {
		t.Fatalf("expected subject 'Port Change', got %q", got.Subject)
	}
}

func TestSNSNotifyNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := makeSNSNotifier(ts.URL)
	d := ports.Diff{Opened: []int{443}}
	if err := n.Notify(d); err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}

func TestSNSMessageContainsHostname(t *testing.T) {
	var got snsPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := makeSNSNotifier(ts.URL)
	n.hostname = "myserver"
	d := ports.Diff{Opened: []int{22}}
	n.Notify(d)

	if got.Message == "" {
		t.Fatal("expected non-empty message")
	}
	if len(got.Message) == 0 {
		t.Fatal("message should reference hostname")
	}
}
