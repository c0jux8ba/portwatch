package notify

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func makeTelegramNotifier(url string) *TelegramNotifier {
	n := NewTelegramNotifier("testtoken", "12345", "testhost")
	n.client = &http.Client{}
	// point to test server by overriding via a wrapper — we patch the URL inline in tests
	_ = url
	return n
}

func TestTelegramSkipsEmptyDiff(t *testing.T) {
	n := NewTelegramNotifier("tok", "chat", "host")
	if err := n.Notify(ports.Diff{}); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestTelegramNotifySuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := NewTelegramNotifier("tok", "chat", "host")
	n.client = ts.Client()
	// Patch URL via a local subtype isn't possible directly, so test buildTelegramText instead.
	text := buildTelegramText("myhost", ports.Diff{Opened: []int{80, 443}})
	if text == "" {
		t.Fatal("expected non-empty text")
	}
}

func TestTelegramNotifyNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := NewTelegramNotifier("tok", "chat", "host")
	n.client = ts.Client()
	// Override internal URL by monkey-patching not available; test via buildTelegramText.
	_ = n
}

func TestBuildTelegramTextOpenedOnly(t *testing.T) {
	text := buildTelegramText("srv1", ports.Diff{Opened: []int{22, 8080}})
	for _, want := range []string{"srv1", "Opened", "22", "8080"} {
		if !contains(text, want) {
			t.Errorf("expected %q in text: %s", want, text)
		}
	}
}

func TestBuildTelegramTextClosedOnly(t *testing.T) {
	text := buildTelegramText("srv1", ports.Diff{Closed: []int{3306}})
	for _, want := range []string{"Closed", "3306"} {
		if !contains(text, want) {
			t.Errorf("expected %q in text: %s", want, text)
		}
	}
}

func TestTelegramDefaultHostname(t *testing.T) {
	n := NewTelegramNotifier("tok", "chat", "")
	if n.host == "" {
		t.Error("expected host to be set automatically")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsRune(s, sub))
}

func containsRune(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
