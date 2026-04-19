package daemon

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func TestNtfyRunnerSkipsEmptyDiff(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	stub := &stubRunner{diff: ports.Diff{}}
	r := NewNtfyRunner(stub, ts.URL, "test")
	_, err := r.Tick()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected ntfy server not to be called for empty diff")
	}
}

func TestNtfyRunnerNotifiesChange(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	stub := &stubRunner{diff: ports.Diff{Opened: []int{9090}}}
	r := NewNtfyRunner(stub, ts.URL, "test")
	_, err := r.Tick()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected ntfy server to be called")
	}
}

func TestNtfyRunnerPropagatesInnerError(t *testing.T) {
	stub := &stubRunner{err: errStub}
	r := NewNtfyRunner(stub, "http://unused", "test")
	_, err := r.Tick()
	if err == nil {
		t.Fatal("expected inner error to propagate")
	}
}
