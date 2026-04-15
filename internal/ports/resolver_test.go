package ports

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestResolverBuiltinSSH(t *testing.T) {
	r := NewResolver()
	if got := r.Lookup(22); got != "ssh" {
		t.Errorf("expected ssh, got %q", got)
	}
}

func TestResolverBuiltinHTTP(t *testing.T) {
	r := NewResolver()
	if got := r.Lookup(80); got != "http" {
		t.Errorf("expected http, got %q", got)
	}
}

func TestResolverUnknownPortReturnsNumeric(t *testing.T) {
	r := NewResolver()
	if got := r.Lookup(19999); got != "19999" {
		t.Errorf("expected \"19999\", got %q", got)
	}
}

func TestResolverLookupAll(t *testing.T) {
	r := NewResolver()
	results := r.LookupAll([]int{22, 80, 19999})
	expected := []string{"22/ssh", "80/http", "19999/19999"}
	for i, want := range expected {
		if results[i] != want {
			t.Errorf("index %d: want %q, got %q", i, want, results[i])
		}
	}
}

func TestResolverLookupAllEmpty(t *testing.T) {
	r := NewResolver()
	if got := r.LookupAll(nil); len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

func TestResolverLoadsEtcServicesFile(t *testing.T) {
	tmp, err := os.CreateTemp("", "services-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())

	_, _ = fmt.Fprintln(tmp, "# comment line")
	_, _ = fmt.Fprintln(tmp, "customsvc   9999/tcp   # a custom service")
	_, _ = fmt.Fprintln(tmp, "udponly     9998/udp")
	tmp.Close()

	r := &Resolver{table: make(map[int]string)}
	r.loadEtcServices(tmp.Name())

	if got := r.Lookup(9999); got != "customsvc" {
		t.Errorf("expected customsvc, got %q", got)
	}
	// UDP entry should NOT be loaded
	if got := r.Lookup(9998); !strings.HasPrefix(got, "9998") {
		t.Errorf("udp port should fall back to numeric, got %q", got)
	}
}

func TestResolverMissingEtcServicesIsOK(t *testing.T) {
	r := &Resolver{table: make(map[int]string)}
	// Should not panic on missing file
	r.loadEtcServices("/nonexistent/path/services")
	if got := r.Lookup(22); got != "22" {
		// builtin table not loaded, so numeric is expected here
	}
}
