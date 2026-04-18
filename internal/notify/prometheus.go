package notify

import (
	"fmt"
	"net/http"

	"github.com/user/portwatch/internal/ports"
)

// PrometheusNotifier exposes port change metrics via an HTTP endpoint
// compatible with Prometheus scraping.
type PrometheusNotifier struct {
	opened int64
	closed int64
	total  int64
	mux    *http.ServeMux
}

// NewPrometheusNotifier creates a PrometheusNotifier and registers its
// /metrics handler on mux. If mux is nil, http.DefaultServeMux is used.
func NewPrometheusNotifier(mux *http.ServeMux) *PrometheusNotifier {
	if mux == nil {
		mux = http.DefaultServeMux
	}
	p := &PrometheusNotifier{mux: mux}
	mux.HandleFunc("/metrics", p.handleMetrics)
	return p
}

// Notify updates internal counters based on the diff.
func (p *PrometheusNotifier) Notify(d ports.Diff) error {
	if d.IsEmpty() {
		return nil
	}
	p.opened += int64(len(d.Opened))
	p.closed += int64(len(d.Closed))
	p.total++
	return nil
}

func (p *PrometheusNotifier) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "# HELP portwatch_opened_total Ports newly opened since start\n")
	fmt.Fprintf(w, "# TYPE portwatch_opened_total counter\n")
	fmt.Fprintf(w, "portwatch_opened_total %d\n", p.opened)
	fmt.Fprintf(w, "# HELP portwatch_closed_total Ports newly closed since start\n")
	fmt.Fprintf(w, "# TYPE portwatch_closed_total counter\n")
	fmt.Fprintf(w, "portwatch_closed_total %d\n", p.closed)
	fmt.Fprintf(w, "# HELP portwatch_events_total Total diff events processed\n")
	fmt.Fprintf(w, "# TYPE portwatch_events_total counter\n")
	fmt.Fprintf(w, "portwatch_events_total %d\n", p.total)
}
