package ports

import (
	"testing"
)

func TestEnricherKnownService(t *testing.T) {
	r := NewResolver(nil)
	e := NewEnricher(r, nil)
	infos := e.Enrich([]int{22, 80})
	if len(infos) != 2 {
		t.Fatalf("expected 2 infos, got %d", len(infos))
	}
	if infos[0].Service != "ssh" {
		t.Errorf("expected ssh, got %s", infos[0].Service)
	}
	if infos[1].Service != "http" {
		t.Errorf("expected http, got %s", infos[1].Service)
	}
}

func TestEnricherUnknownService(t *testing.T) {
	r := NewResolver(nil)
	e := NewEnricher(r, nil)
	infos := e.Enrich([]int{19999})
	if len(infos) != 1 {
		t.Fatalf("expected 1 info, got %d", len(infos))
	}
	if infos[0].Service != "19999" {
		t.Errorf("expected numeric fallback, got %s", infos[0].Service)
	}
}

func TestEnricherNilResolver(t *testing.T) {
	e := NewEnricher(nil, nil)
	infos := e.Enrich([]int{443})
	if len(infos) != 1 {
		t.Fatalf("expected 1 info, got %d", len(infos))
	}
	if infos[0].Service != "" {
		t.Errorf("expected empty service with nil resolver, got %s", infos[0].Service)
	}
}

func TestEnricherEmptyPorts(t *testing.T) {
	r := NewResolver(nil)
	e := NewEnricher(r, nil)
	infos := e.Enrich([]int{})
	if len(infos) != 0 {
		t.Fatalf("expected 0 infos, got %d", len(infos))
	}
}

func TestEnricherPortInfoFields(t *testing.T) {
	r := NewResolver(nil)
	e := NewEnricher(r, nil)
	infos := e.Enrich([]int{443})
	if infos[0].Port != 443 {
		t.Errorf("expected port 443, got %d", infos[0].Port)
	}
	if infos[0].PID != 0 {
		t.Errorf("expected PID 0 with no process resolver, got %d", infos[0].PID)
	}
	if infos[0].Process != "" {
		t.Errorf("expected empty process name, got %s", infos[0].Process)
	}
}
