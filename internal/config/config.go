package config

import (
	"encoding/json"
	"os"
	"time"
)

// Config holds all portwatch runtime settings.
type Config struct {
	// PortRange is the inclusive [low, high] range to scan.
	PortRange [2]int `json:"port_range"`
	// IntervalSeconds between each scan cycle.
	IntervalSeconds int `json:"interval_seconds"`
	// WebhookURL is an optional HTTP endpoint to POST change events to.
	WebhookURL string `json:"webhook_url"`
	// DesktopNotify enables OS desktop notifications.
	DesktopNotify bool `json:"desktop_notify"`
	// AppName is the name shown in desktop notifications.
	AppName string `json:"app_name"`
	// HistoryPath is where scan history is persisted across restarts.
	HistoryPath string `json:"history_path"`
	// MaxHistory is the maximum number of snapshots to retain.
	MaxHistory int `json:"max_history"`
	// ExcludePorts lists individual ports to ignore during scans.
	ExcludePorts []int `json:"exclude_ports"`
	// ExcludeRanges lists [low, high] port ranges to ignore.
	ExcludeRanges [][]int `json:"exclude_ranges"`
}

// Interval converts IntervalSeconds to a time.Duration.
func (c *Config) Interval() time.Duration {
	return time.Duration(c.IntervalSeconds) * time.Second
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		PortRange:       [2]int{1, 65535},
		IntervalSeconds: 60,
		DesktopNotify:   false,
		AppName:         "portwatch",
		HistoryPath:     "/tmp/portwatch_history.json",
		MaxHistory:      50,
		ExcludePorts:    []int{},
		ExcludeRanges:   [][]int{},
	}
}

// Load reads a JSON config file from path, falling back to DefaultConfig
// when the file does not exist. Any other error is returned to the caller.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
