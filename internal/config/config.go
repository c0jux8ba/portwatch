package config

import (
	"encoding/json"
	"errors"
	"os"
)

// Config holds all portwatch runtime settings.
type Config struct {
	PortRangeStart  int    `json:"port_range_start"`
	PortRangeEnd    int    `json:"port_range_end"`
	IntervalSeconds int    `json:"interval_seconds"`
	WebhookURL      string `json:"webhook_url"`
	AppName         string `json:"app_name"`
	DesktopNotify   bool   `json:"desktop_notify"`
	HistoryFile     string `json:"history_file"`
	HistoryMaxLen   int    `json:"history_max_len"`
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() Config {
	return Config{
		PortRangeStart:  1,
		PortRangeEnd:    65535,
		IntervalSeconds: 30,
		AppName:         "portwatch",
		DesktopNotify:   false,
		HistoryFile:     "portwatch_history.json",
		HistoryMaxLen:   200,
	}
}

// Load reads a JSON config file, falling back to defaults for missing fields.
func Load(path string) (Config, error) {
	cfg := DefaultConfig()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, cfg.Validate()
}

// Validate checks that the Config fields are logically consistent.
func (c Config) Validate() error {
	if c.PortRangeStart < 1 || c.PortRangeStart > 65535 {
		return errors.New("port_range_start must be between 1 and 65535")
	}
	if c.PortRangeEnd < 1 || c.PortRangeEnd > 65535 {
		return errors.New("port_range_end must be between 1 and 65535")
	}
	if c.PortRangeStart > c.PortRangeEnd {
		return errors.New("port_range_start must be <= port_range_end")
	}
	if c.IntervalSeconds < 1 {
		return errors.New("interval_seconds must be >= 1")
	}
	if c.HistoryMaxLen < 1 {
		return errors.New("history_max_len must be >= 1")
	}
	return nil
}
