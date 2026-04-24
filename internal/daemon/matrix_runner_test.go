package daemon

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func TestMatrixRunnerRequiresFields(t *testing.T) {
	cases := []struct {
		homeserver, roomID, token string
	}{
		{"", "!r:s", "tok"},
		{"http://x", "", "tok"},
		{"http://x", "!r:s", ""},
	}
	for _, c := range cases {
		_, err := NewMatrixRunner(nil, c.homeserver, c.roomID, c.token, "pw")
		if err == nil {
			t.Errorf("expected error for homeserver=%q roomID=%q token=%q", c.homeserver, c.roomID, c.token)
		}
	}
}

func TestMatrixRunnerTickWithDiffSkipsEmpty(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not have been called for empty diff")
	}))
	defer ts.Close()

	r, err := NewMatrixRunner(&noopRunner{}, ts.URL, "!room:server", "tok", "pw")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := r.TickWithDiff(ports.Diff{}); err != nil {
		t.Fatalf("expected nil for empty diff, got %v", err)
	}
}

func TestMatrixRunnerTickWithDiffNotifies(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"event_id":"$ok"}`))
	}))
	defer ts.Close()

	r, err := NewMatrixRunner(&noopRunner{}, ts.URL, "!room:server", "tok", "pw")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := r.TickWithDiff(ports.Diff{Opened: []int{8080}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected Matrix endpoint to be called")
	}
}

func TestMatrixRunnerTickDelegates(t *testing.T) {
	sentinel := errors.New("inner error")
	inner := &errRunner{err: sentinel}

	r, _ := NewMatrixRunner(inner, "http://x", "!r:s", "tok", "pw")
	if err := r.Tick(); !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

// noopRunner satisfies the Runner interface without doing anything.
type noopRunner struct{}

func (n *noopRunner) Tick() error { return nil }

// errRunner always returns the configured error from Tick.
type errRunner struct{ err error }

func (e *errRunner) Tick() error { return e.err }
