package config

import (
	"encoding/json"
	"os"
)

// Config holds the full portwatch configuration.
type Config struct {
	StartPort   int      `json:"start_port"`
	EndPort     int      `json:"end_port"`
	IntervalSec int      `json:"interval_sec"`
	WebhookURL  string   `json:"webhook_url"`
	SlackURL    string   `json:"slack_url"`
	DesktopApp  string   `json:"desktop_app"`
	Exclude     []string `json:"exclude"`
	CooldownSec int      `json:"cooldown_sec"`
	MaxHistory  int      `json:"max_history"`
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() Config {
	return Config{
		StartPort:   1,
		EndPort:     65535,
		IntervalSec: 30,
		CooldownSec: 60,
		MaxHistory:  100,
	}
}

// Load reads a JSON config file from path. Missing file returns DefaultConfig.
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

	return cfg, nil
}
