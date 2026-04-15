package ports

import (
	"encoding/json"
	"os"
	"time"
)

// PortEvent records a single change event for persistence.
type PortEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Opened    []int     `json:"opened,omitempty"`
	Closed    []int     `json:"closed,omitempty"`
}

// History manages a rolling log of port change events.
type History struct {
	path   string
	events []PortEvent
	maxLen int
}

// NewHistory creates a History backed by the given file path.
// Existing events are loaded if the file exists.
func NewHistory(path string, maxLen int) (*History, error) {
	h := &History{path: path, maxLen: maxLen}
	if err := h.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return h, nil
}

// Record appends a new event and persists the history file.
func (h *History) Record(opened, closed []int) error {
	if len(opened) == 0 && len(closed) == 0 {
		return nil
	}
	e := PortEvent{
		Timestamp: time.Now().UTC(),
		Opened:    opened,
		Closed:    closed,
	}
	h.events = append(h.events, e)
	if len(h.events) > h.maxLen {
		h.events = h.events[len(h.events)-h.maxLen:]
	}
	return h.save()
}

// Events returns a copy of all stored events.
func (h *History) Events() []PortEvent {
	out := make([]PortEvent, len(h.events))
	copy(out, h.events)
	return out
}

func (h *History) load() error {
	data, err := os.ReadFile(h.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &h.events)
}

func (h *History) save() error {
	data, err := json.MarshalIndent(h.events, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(h.path, data, 0o644)
}
