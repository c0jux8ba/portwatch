package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "portwatch-*.json")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoadMissingFileReturnsDefault(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "nonexistent.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PortRange.Start != 1 || cfg.PortRange.End != 65535 {
		t.Errorf("unexpected default port range: %+v", cfg.PortRange)
	}
	if cfg.Interval.Duration != 30*time.Second {
		t.Errorf("unexpected default interval: %v", cfg.Interval.Duration)
	}
}

func TestLoadValidConfig(t *testing.T) {
	raw := `{"port_range":{"start":1024,"end":9000},"interval":"1m","webhook_url":"http://example.com","desktop_notify":false}`
	path := writeTemp(t, raw)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PortRange.Start != 1024 || cfg.PortRange.End != 9000 {
		t.Errorf("unexpected port range: %+v", cfg.PortRange)
	}
	if cfg.Interval.Duration != time.Minute {
		t.Errorf("unexpected interval: %v", cfg.Interval.Duration)
	}
	if cfg.WebhookURL != "http://example.com" {
		t.Errorf("unexpected webhook_url: %v", cfg.WebhookURL)
	}
	if cfg.DesktopNotify {
		t.Error("expected desktop_notify to be false")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	path := writeTemp(t, `{not valid json}`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestValidatePortRangeInverted(t *testing.T) {
	cfg := DefaultConfig()
	cfg.PortRange = PortRange{Start: 9000, End: 1024}
	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for inverted port range")
	}
}

func TestValidateNegativeInterval(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Interval = Duration{-time.Second}
	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for negative interval")
	}
}

func TestDurationRoundTrip(t *testing.T) {
	raw := `{"interval":"45s"}`
	path := writeTemp(t, raw)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Interval.Duration != 45*time.Second {
		t.Errorf("expected 45s, got %v", cfg.Interval.Duration)
	}
}
