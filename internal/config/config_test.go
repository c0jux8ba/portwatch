package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoadMissingFileReturnsDefault(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.json")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	def := DefaultConfig()
	if cfg.IntervalSeconds != def.IntervalSeconds {
		t.Errorf("expected default interval %d, got %d", def.IntervalSeconds, cfg.IntervalSeconds)
	}
}

func TestLoadValidConfig(t *testing.T) {
	p := writeTemp(t, `{"port_range_start":1024,"port_range_end":2048,"interval_seconds":10,"history_max_len":50}`)
	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PortRangeStart != 1024 {
		t.Errorf("expected 1024, got %d", cfg.PortRangeStart)
	}
	if cfg.HistoryMaxLen != 50 {
		t.Errorf("expected history_max_len 50, got %d", cfg.HistoryMaxLen)
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	p := writeTemp(t, `{not valid json`)
	_, err := Load(p)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestValidatePortRangeInverted(t *testing.T) {
	cfg := DefaultConfig()
	cfg.PortRangeStart = 9000
	cfg.PortRangeEnd = 1000
	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for inverted range")
	}
}

func TestValidateIntervalZero(t *testing.T) {
	cfg := DefaultConfig()
	cfg.IntervalSeconds = 0
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for zero interval")
	}
}

func TestValidateHistoryMaxLenZero(t *testing.T) {
	cfg := DefaultConfig()
	cfg.HistoryMaxLen = 0
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for zero history_max_len")
	}
}

func TestDefaultHistoryFields(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.HistoryFile == "" {
		t.Error("expected non-empty default HistoryFile")
	}
	if cfg.HistoryMaxLen <= 0 {
		t.Errorf("expected positive default HistoryMaxLen, got %d", cfg.HistoryMaxLen)
	}
}
