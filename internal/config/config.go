package config

import (
	"encoding/json"
	"os"
	"time"
)

// Config holds portwatch runtime configuration.
type Config struct {
	PortRange        string        `json:"port_range"`
	ScanInterval     time.Duration `json:"scan_interval"`
	WebhookURL       string        `json:"webhook_url"`
	DesktopNotify    bool          `json:"desktop_notify"`
	AppName          string        `json:"app_name"`
	BaselinePath     string        `json:"baseline_path"`
	HistoryPath      string        `json:"history_path"`
	HistoryMaxLen    int           `json:"history_max_len"`
	CooldownSeconds  int           `json:"cooldown_seconds"`
	RetryAttempts    int           `json:"retry_attempts"`
	RetryDelayMs     int           `json:"retry_delay_ms"`
	ExcludePorts     []string      `json:"exclude_ports"`
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() Config {
	return Config{
		PortRange:       "1-65535",
		ScanInterval:    30 * time.Second,
		DesktopNotify:   true,
		AppName:         "portwatch",
		BaselinePath:    ".portwatch_baseline.json",
		HistoryPath:     ".portwatch_history.json",
		HistoryMaxLen:   100,
		CooldownSeconds: 60,
		RetryAttempts:   3,
		RetryDelayMs:    500,
	}
}

// Load reads a JSON config file, falling back to defaults for missing fields.
func Load(path string) (Config, error) {
	cfg := DefaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
