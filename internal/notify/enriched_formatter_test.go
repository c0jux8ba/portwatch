package notify

import (
	"strings"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func makeEnrichedFormatter() *EnrichedFormatter {
	r := ports.NewResolver(nil)
	e := ports.NewEnricher(r, nil)
	return NewEnrichedFormatter(e, "testhost")
}

func TestEnrichedFormatterOpened(t *testing.T) {
	f := makeEnrichedFormatter()
	d := ports.Diff{Opened: []int{22}, Closed: []int{}}
	out := f.Format(d)
	if !strings.Contains(out, "Opened") {
		t.Errorf("expected 'Opened' in output, got: %s", out)
	}
	if !strings.Contains(out, ":22") {
		t.Errorf("expected ':22' in output, got: %s", out)
	}
	if !strings.Contains(out, "ssh") {
		t.Errorf("expected 'ssh' in output, got: %s", out)
	}
}

func TestEnrichedFormatterClosed(t *testing.T) {
	f := makeEnrichedFormatter()
	d := ports.Diff{Opened: []int{}, Closed: []int{80}}
	out := f.Format(d)
	if !strings.Contains(out, "Closed") {
		t.Errorf("expected 'Closed' in output, got: %s", out)
	}
	if !strings.Contains(out, "http") {
		t.Errorf("expected 'http' in output, got: %s", out)
	}
}

func TestEnrichedFormatterHostname(t *testing.T) {
	f := makeEnrichedFormatter()
	d := ports.Diff{Opened: []int{443}, Closed: []int{}}
	out := f.Format(d)
	if !strings.Contains(out, "testhost") {
		t.Errorf("expected hostname in output, got: %s", out)
	}
}

func TestEnrichedFormatterDefaultHostname(t *testing.T) {
	r := ports.NewResolver(nil)
	e := ports.NewEnricher(r, nil)
	f := NewEnrichedFormatter(e, "")
	if f.hostname != "localhost" {
		t.Errorf("expected default hostname 'localhost', got %s", f.hostname)
	}
}

func TestEnrichedFormatterUnknownPort(t *testing.T) {
	f := makeEnrichedFormatter()
	d := ports.Diff{Opened: []int{19999}, Closed: []int{}}
	out := f.Format(d)
	if !strings.Contains(out, "19999") {
		t.Errorf("expected numeric port in output, got: %s", out)
	}
}
