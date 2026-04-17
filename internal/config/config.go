package config

import (
	"encoding/json"
	"os"
)

// Config holds portwatch runtime configuration.
type Config struct {
	IntervalSeconds  int      `json:"interval_seconds"`
	PortRange        string   `json:"port_range"`
	ExcludePorts     []int    `json:"exclude_ports"`
	WebhookURL       string   `json:"webhook_url"`
	SlackWebhookURL  string   `json:"slack_webhook_url"`
	PagerDutyKey     string   `json:"pagerduty_key"`
	EmailSMTP        string   `json:"email_smtp"`
	EmailFrom        string   `json:"email_from"`
	EmailTo          string   `json:"email_to"`
	DesktopNotify    bool     `json:"desktop_notify"`
	ConsoleNotify    bool     `json:"console_notify"`
	ConsolePrefix    string   `json:"console_prefix"`
	BaselinePath     string   `json:"baseline_path"`
	HistoryPath      string   `json:"history_path"`
	CooldownSeconds  int      `json:"cooldown_seconds"`
	MaxHistoryLen    int      `json:"max_history_len"`
}

func DefaultConfig() Config {
	return Config{
		IntervalSeconds: 30,
		PortRange:       "1-65535",
		DesktopNotify:   false,
		ConsoleNotify:   true,
		ConsolePrefix:   "portwatch",
		BaselinePath:    ".portwatch_baseline.json",
		HistoryPath:     ".portwatch_history.json",
		CooldownSeconds: 60,
		MaxHistoryLen:   100,
	}
}

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
