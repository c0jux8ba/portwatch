package notify

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func makePrometheusNotifier() (*PrometheusNotifier, *http.ServeMux) {
	mux := http.NewServeMux()
	p := NewPrometheusNotifier(mux)
	return p, mux
}

func TestPrometheusSkipsEmptyDiff(t *testing.T) {
	p, _ := makePrometheusNotifier()
	if err := p.Notify(ports.Diff{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.total != 0 {
		t.Errorf("expected total=0, got %d", p.total)
	}
}

func TestPrometheusCountsOpened(t *testing.T) {
	p, _ := makePrometheusNotifier()
	_ = p.Notify(ports.Diff{Opened: []int{80, 443}})
	if p.opened != 2 {
		t.Errorf("expected opened=2, got %d", p.opened)
	}
	if p.total != 1 {
		t.Errorf("expected total=1, got %d", p.total)
	}
}

func TestPrometheusCountsClosed(t *testing.T) {
	p, _ := makePrometheusNotifier()
	_ = p.Notify(ports.Diff{Closed: []int{22}})
	if p.closed != 1 {
		t.Errorf("expected closed=1, got %d", p.closed)
	}
}

func TestPrometheusMetricsEndpoint(t *testing.T) {
	p, mux := makePrometheusNotifier()
	_ = p.Notify(ports.Diff{Opened: []int{8080}, Closed: []int{22, 23}})

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	body := rec.Body.String()
	for _, want := range []string{
		"portwatch_opened_total 1",
		"portwatch_closed_total 2",
		"portwatch_events_total 1",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("metrics body missing %q\ngot:\n%s", want, body)
		}
	}
}

func TestPrometheusDefaultMux(t *testing.T) {
	// Should not panic when nil mux is passed
	p := NewPrometheusNotifier(nil)
	if p == nil {
		t.Fatal("expected non-nil notifier")
	}
}
