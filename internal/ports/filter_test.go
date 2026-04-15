package ports

import (
	"testing"
)

func TestFilterNilAllowsAll(t *testing.T) {
	var f *Filter
	for _, p := range []int{1, 80, 443, 65535} {
		if !f.Allowed(p) {
			t.Errorf("nil filter should allow port %d", p)
		}
	}
}

func TestFilterExcludesSinglePorts(t *testing.T) {
	f := NewFilter([]int{22, 8080}, nil)
	if f.Allowed(22) {
		t.Error("port 22 should be excluded")
	}
	if f.Allowed(8080) {
		t.Error("port 8080 should be excluded")
	}
	if !f.Allowed(80) {
		t.Error("port 80 should be allowed")
	}
}

func TestFilterExcludesRange(t *testing.T) {
	f := NewFilter(nil, [][]int{{3000, 3010}})
	for p := 3000; p <= 3010; p++ {
		if f.Allowed(p) {
			t.Errorf("port %d inside range should be excluded", p)
		}
	}
	if !f.Allowed(2999) {
		t.Error("port 2999 should be allowed")
	}
	if !f.Allowed(3011) {
		t.Error("port 3011 should be allowed")
	}
}

func TestFilterInvalidRangeIgnored(t *testing.T) {
	// high < low — should be silently ignored
	f := NewFilter(nil, [][]int{{9000, 8000}})
	if !f.Allowed(8500) {
		t.Error("invalid range should be ignored; port 8500 should be allowed")
	}
}

func TestFilterApply(t *testing.T) {
	f := NewFilter([]int{22}, [][]int{{8000, 8002}})
	input := []int{22, 80, 443, 8000, 8001, 8002, 9090}
	got := f.Apply(input)
	want := []int{80, 443, 9090}
	if len(got) != len(want) {
		t.Fatalf("Apply: got %v, want %v", got, want)
	}
	for i, p := range want {
		if got[i] != p {
			t.Errorf("Apply[%d]: got %d, want %d", i, got[i], p)
		}
	}
}

func TestFilterApplyNilFilter(t *testing.T) {
	var f *Filter
	input := []int{22, 80, 443}
	got := f.Apply(input)
	if len(got) != len(input) {
		t.Fatalf("nil Apply should return all ports, got %v", got)
	}
}
