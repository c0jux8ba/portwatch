package ports

// Filter holds configuration for excluding ports or ranges from scan results.
type Filter struct {
	// ExcludePorts is a set of individual ports to ignore.
	excludePorts map[int]struct{}
	// ExcludeRanges holds [low, high] inclusive pairs.
	excludeRanges [][2]int
}

// NewFilter builds a Filter from a list of excluded ports and ranges.
// ranges must be provided as pairs: [][]int{{low1, high1}, {low2, high2}}.
func NewFilter(excludePorts []int, excludeRanges [][]int) *Filter {
	f := &Filter{
		excludePorts: make(map[int]struct{}, len(excludePorts)),
	}
	for _, p := range excludePorts {
		f.excludePorts[p] = struct{}{}
	}
	for _, r := range excludeRanges {
		if len(r) == 2 && r[0] <= r[1] {
			f.excludeRanges = append(f.excludeRanges, [2]int{r[0], r[1]})
		}
	}
	return f
}

// Allowed returns true when port p should be included in results.
func (f *Filter) Allowed(p int) bool {
	if f == nil {
		return true
	}
	if _, excluded := f.excludePorts[p]; excluded {
		return false
	}
	for _, r := range f.excludeRanges {
		if p >= r[0] && p <= r[1] {
			return false
		}
	}
	return true
}

// Apply returns only the ports from ps that pass the filter.
func (f *Filter) Apply(ps []int) []int {
	if f == nil {
		return ps
	}
	out := make([]int, 0, len(ps))
	for _, p := range ps {
		if f.Allowed(p) {
			out = append(out, p)
		}
	}
	return out
}
