package notify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func makePDNotifier(url string) *pagerDutyNotifier {
	n := NewPagerDutyNotifier("test-key").(*pagerDutyNotifier)
	n.client = &http.Client{}
	// override URL via a small wrapper — we redirect via transport
	_ = url
	return n
}

func TestPagerDutySkipsEmptyDiff(t *testing.T) {
	n := NewPagerDutyNotifier("key")
	if err := n.Notify(ports.Diff{}); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestPagerDutySuccess(t *testing.T) {
	var received pdPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(202)
	}))
	defer ts.Close()

	n := &pagerDutyNotifier{
		integrationKey: "my-key",
		client:         ts.Client(),
		formatter:      NewFormatter("testhost"),
	}
	// point at test server
	orig := pagerDutyEventURL
	pagerDutyEventURLVar = ts.URL
	defer func() { pagerDutyEventURLVar = orig }()

	diff := ports.Diff{Opened: []int{9200}}
	if err := n.notifyTo(diff, ts.URL); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.RoutingKey != "my-key" {
		t.Errorf("expected routing key my-key, got %s", received.RoutingKey)
	}
	if received.EventAction != "trigger" {
		t.Errorf("expected trigger, got %s", received.EventAction)
	}
}

func TestPagerDutyNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
	}))
	defer ts.Close()

	n := &pagerDutyNotifier{
		integrationKey: "key",
		client:         ts.Client(),
		formatter:      NewFormatter(""),
	}
	diff := ports.Diff{Opened: []int{8080}}
	if err := n.notifyTo(diff, ts.URL); err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}
