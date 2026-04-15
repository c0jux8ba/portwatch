package ports

import (
	"reflect"
	"testing"
)

func TestCompareNoChange(t *testing.T) {
	prev := []int{80, 443, 8080}
	curr := []int{80, 443, 8080}
	d := Compare(prev, curr)
	if !d.IsEmpty() {
		t.Errorf("expected empty diff, got opened=%v closed=%v", d.Opened, d.Closed)
	}
}

func TestCompareDetectsOpened(t *testing.T) {
	prev := []int{80}
	curr := []int{80, 443, 8080}
	d := Compare(prev, curr)
	want := []int{443, 8080}
	if !reflect.DeepEqual(d.Opened, want) {
		t.Errorf("opened: got %v, want %v", d.Opened, want)
	}
	if len(d.Closed) != 0 {
		t.Errorf("expected no closed ports, got %v", d.Closed)
	}
}

func TestCompareDetectsClosed(t *testing.T) {
	prev := []int{80, 443, 8080}
	curr := []int{80}
	d := Compare(prev, curr)
	want := []int{443, 8080}
	if !reflect.DeepEqual(d.Closed, want) {
		t.Errorf("closed: got %v, want %v", d.Closed, want)
	}
	if len(d.Opened) != 0 {
		t.Errorf("expected no opened ports, got %v", d.Opened)
	}
}

func TestCompareBothDirections(t *testing.T) {
	prev := []int{80, 443}
	curr := []int{80, 8080}
	d := Compare(prev, curr)
	if !reflect.DeepEqual(d.Opened, []int{8080}) {
		t.Errorf("opened: got %v, want [8080]", d.Opened)
	}
	if !reflect.DeepEqual(d.Closed, []int{443}) {
		t.Errorf("closed: got %v, want [443]", d.Closed)
	}
}

func TestIsEmptyFalseWhenChanged(t *testing.T) {
	d := Diff{Opened: []int{9000}}
	if d.IsEmpty() {
		t.Error("expected IsEmpty to return false")
	}
}

func TestItoa(t *testing.T) {
	if itoa(8080) != "8080" {
		t.Errorf("itoa(8080) = %q, want \"8080\"", itoa(8080))
	}
}
