package notify

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func makeNtfyNotifier(url string) *NtfyNotifier {
	n := NewNtfyNotifier(url, "portwatch")
	n.hostname = "testhost"
	return n
}

func TestNtfySkipsEmptyDiff(t *testing.T) {
	n := makeNtfyNotifier("http://unused")
	if err := n.Notify(ports.Diff{}); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestNtfyNotifySuccess(t *testing.T) {
	var gotBody, gotTitle string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTitle = r.Header.Get("Title")
		buf := new(strings.Builder)
		buf.ReadFrom(r.Body)
		gotBody = buf.String()
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := makeNtfyNotifier(ts.URL)
	err := n.Notify(ports.Diff{Opened: []int{8080}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(gotBody, "8080") {
		t.Errorf("expected body to contain port, got: %s", gotBody)
	}
	if !strings.Contains(gotTitle, "testhost") {
		t.Errorf("expected title to contain hostname, got: %s", gotTitle)
	}
}

func TestNtfyNotifyNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	n := makeNtfyNotifier(ts.URL)
	err := n.Notify(ports.Diff{Opened: []int{443}})
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestBuildNtfyMessageOpenedOnly(t *testing.T) {
	msg := buildNtfyMessage(ports.Diff{Opened: []int{80, 443}}, "myhost")
	if !strings.Contains(msg, "Opened") || !strings.Contains(msg, "myhost") {
		t.Errorf("unexpected message: %s", msg)
	}
	if strings.Contains(msg, "Closed") {
		t.Errorf("should not contain Closed section")
	}
}

func TestBuildNtfyMessageBothDirections(t *testing.T) {
	msg := buildNtfyMessage(ports.Diff{Opened: []int{8080}, Closed: []int{22}}, "host")
	if !strings.Contains(msg, "Opened") || !strings.Contains(msg, "Closed") {
		t.Errorf("expected both sections, got: %s", msg)
	}
}
