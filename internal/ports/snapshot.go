package ports

import "sort"

// Snapshot holds an immutable, sorted list of open ports captured at a
// single point in time.
type Snapshot struct {
	Ports []int
}

// NewSnapshot creates a Snapshot from the provided port list, deduplicating
// and sorting the values for consistent comparison.
func NewSnapshot(ports []int) Snapshot {
	seen := make(map[int]bool, len(ports))
	uniq := make([]int, 0, len(ports))
	for _, p := range ports {
		if !seen[p] {
			seen[p] = true
			uniq = append(uniq, p)
		}
	}
	sort.Ints(uniq)
	return Snapshot{Ports: uniq}
}

// DiffFrom returns a Diff between this snapshot and a newer one.
func (s Snapshot) DiffFrom(newer Snapshot) Diff {
	return Compare(s.Ports, newer.Ports)
}

// Equal reports whether two snapshots contain the same ports.
func (s Snapshot) Equal(other Snapshot) bool {
	if len(s.Ports) != len(other.Ports) {
		return false
	}
	for i := range s.Ports {
		if s.Ports[i] != other.Ports[i] {
			return false
		}
	}
	return true
}
