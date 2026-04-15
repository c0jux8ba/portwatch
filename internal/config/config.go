package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config holds the portwatch daemon configuration.
type Config struct {
	PortRange    PortRange     `json:"port_range"`
	Interval     Duration      `json:"interval"`
	WebhookURL   string        `json:"webhook_url,omitempty"`
	DesktopNotify bool         `json:"desktop_notify"`
	AppName      string        `json:"app_name,omitempty"`
}

// PortRange defines the start and end of the port scan range.
type PortRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// Duration is a wrapper around time.Duration for JSON unmarshalling.
type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	d.Duration = parsed
	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Duration.String())
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		PortRange:     PortRange{Start: 1, End: 65535},
		Interval:      Duration{30 * time.Second},
		DesktopNotify: true,
		AppName:       "portwatch",
	}
}

// Load reads a JSON config file from path. Missing file returns DefaultConfig.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks that the Config fields are sensible.
func (c *Config) Validate() error {
	if c.PortRange.Start < 1 || c.PortRange.Start > 65535 {
		return fmt.Errorf("port_range.start must be between 1 and 65535")
	}
	if c.PortRange.End < 1 || c.PortRange.End > 65535 {
		return fmt.Errorf("port_range.end must be between 1 and 65535")
	}
	if c.PortRange.Start > c.PortRange.End {
		return fmt.Errorf("port_range.start must be <= port_range.end")
	}
	if c.Interval.Duration <= 0 {
		return fmt.Errorf("interval must be positive")
	}
	return nil
}
