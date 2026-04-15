package notify

import (
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func TestDesktopNotifySkipsEmptyDiff(t *testing.T) {
	d := NewDesktopNotifier("testapp")
	// An empty diff must never attempt to spawn a subprocess.
	err := d.Notify(ports.Diff{})
	if err != nil {
		t.Fatalf("expected no error for empty diff, got: %v", err)
	}
}

func TestNewDesktopNotifierDefaultName(t *testing.T) {
	d := NewDesktopNotifier("")
	if d.AppName != "portwatch" {
		t.Errorf("expected default AppName 'portwatch', got %q", d.AppName)
	}
}

func TestNewDesktopNotifierCustomName(t *testing.T) {
	d := NewDesktopNotifier("myapp")
	if d.AppName != "myapp" {
		t.Errorf("expected AppName 'myapp', got %q", d.AppName)
	}
}

func TestBuildBodyOpenedOnly(t *testing.T) {
	diff := ports.Diff{Opened: []int{8080, 9090}}
	body := buildBody(diff)
	expected := "Opened: 8080, 9090"
	if body != expected {
		t.Errorf("expected %q, got %q", expected, body)
	}
}

func TestBuildBodyClosedOnly(t *testing.T) {
	diff := ports.Diff{Closed: []int{22}}
	body := buildBody(diff)
	expected := "Closed: 22"
	if body != expected {
		t.Errorf("expected %q, got %q", expected, body)
	}
}

func TestBuildBodyBoth(t *testing.T) {
	diff := ports.Diff{Opened: []int{443}, Closed: []int{80}}
	body := buildBody(diff)
	expected := "Opened: 443 | Closed: 80"
	if body != expected {
		t.Errorf("expected %q, got %q", expected, body)
	}
}

func TestIntsToStrings(t *testing.T) {
	result := intsToStrings([]int{22, 80, 443})
	if len(result) != 3 || result[0] != "22" || result[1] != "80" || result[2] != "443" {
		t.Errorf("unexpected result: %v", result)
	}
}
