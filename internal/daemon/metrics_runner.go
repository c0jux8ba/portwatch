package daemon

import (
	"fmt"
	"net/http"
	"time"

	"github.com/user/portwatch/internal/notify"
)

// MetricsRunner wraps a PrometheusNotifier and serves its /metrics endpoint
// on the configured address for the lifetime of the daemon.
type MetricsRunner struct {
	notifier *notify.PrometheusNotifier
	server   *http.Server
}

// NewMetricsRunner creates a MetricsRunner listening on addr (e.g. ":9090").
func NewMetricsRunner(addr string) *MetricsRunner {
	mux := http.NewServeMux()
	n := notify.NewPrometheusNotifier(mux)
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	return &MetricsRunner{notifier: n, server: srv}
}

// Notifier returns the underlying PrometheusNotifier so it can be composed
// into a notify.Multi chain.
func (m *MetricsRunner) Notifier() *notify.PrometheusNotifier {
	return m.notifier
}

// Start begins serving metrics in a background goroutine.
// It returns immediately; call Stop to shut down.
func (m *MetricsRunner) Start() error {
	errCh := make(chan error, 1)
	go func() {
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("metrics server: %w", err)
		}
	}()
	select {
	case err := <-errCh:
		return err
	case <-time.After(50 * time.Millisecond):
		return nil
	}
}

// Stop gracefully shuts down the metrics HTTP server.
func (m *MetricsRunner) Stop() error {
	return m.server.Close()
}
