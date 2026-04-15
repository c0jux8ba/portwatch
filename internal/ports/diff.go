package ports

// Diff represents changes between two port snapshots.
type Diff struct {
	Opened []PortState
	Closed []PortState
}

// HasChanges returns true when at least one port opened or closed.
func (d *Diff) HasChanges() bool {
	return len(d.Opened) > 0 || len(d.Closed) > 0
}

// Compare calculates which ports opened or closed between prev and curr.
func Compare(prev, curr *Snapshot) *Diff {
	prevSet := toSet(prev.Ports)
	currSet := toSet(curr.Ports)

	var opened, closed []PortState

	for key, state := range currSet {
		if _, existed := prevSet[key]; !existed {
			opened = append(opened, state)
		}
	}

	for key, state := range prevSet {
		if _, exists := currSet[key]; !exists {
			closed = append(closed, state)
		}
	}

	return &Diff{
		Opened: opened,
		Closed: closed,
	}
}

// toSet converts a slice of PortState into a map keyed by "proto:port".
func toSet(ports []PortState) map[string]PortState {
	m := make(map[string]PortState, len(ports))
	for _, p := range ports {
		key := p.Protocol + ":" + itoa(p.Port)
		m[key] = p
	}
	return m
}

// itoa is a minimal int-to-string helper to avoid importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [10]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}
