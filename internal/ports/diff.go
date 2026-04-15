package ports

import (
	"fmt"
	"sort"
)

// Diff holds the result of comparing two port snapshots.
type Diff struct {
	Opened []int
	Closed []int
}

// IsEmpty returns true when no ports changed.
func (d Diff) IsEmpty() bool {
	return len(d.Opened) == 0 && len(d.Closed) == 0
}

// Compare returns a Diff between the previous and current port sets.
func Compare(previous, current []int) Diff {
	prev := toSet(previous)
	curr := toSet(current)

	var opened, closed []int

	for p := range curr {
		if !prev[p] {
			opened = append(opened, p)
		}
	}
	for p := range prev {
		if !curr[p] {
			closed = append(closed, p)
		}
	}

	sort.Ints(opened)
	sort.Ints(closed)

	return Diff{Opened: opened, Closed: closed}
}

func toSet(ports []int) map[int]bool {
	s := make(map[int]bool, len(ports))
	for _, p := range ports {
		s[p] = true
	}
	return s
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}
