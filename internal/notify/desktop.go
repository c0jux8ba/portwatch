package notify

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/user/portwatch/internal/ports"
)

// DesktopNotifier sends desktop notifications for port changes.
type DesktopNotifier struct {
	AppName string
}

// NewDesktopNotifier creates a new DesktopNotifier with the given app name.
func NewDesktopNotifier(appName string) *DesktopNotifier {
	if appName == "" {
		appName = "portwatch"
	}
	return &DesktopNotifier{AppName: appName}
}

// Notify sends a desktop notification summarising the port diff.
// It is a no-op when the diff contains no changes.
func (d *DesktopNotifier) Notify(diff ports.Diff) error {
	if len(diff.Opened) == 0 && len(diff.Closed) == 0 {
		return nil
	}

	title := fmt.Sprintf("%s: port change detected", d.AppName)
	body := buildBody(diff)

	return sendDesktopNotification(title, body)
}

func buildBody(diff ports.Diff) string {
	var parts []string
	if len(diff.Opened) > 0 {
		parts = append(parts, fmt.Sprintf("Opened: %s", strings.Join(intsToStrings(diff.Opened), ", ")))
	}
	if len(diff.Closed) > 0 {
		parts = append(parts, fmt.Sprintf("Closed: %s", strings.Join(intsToStrings(diff.Closed), ", ")))
	}
	return strings.Join(parts, " | ")
}

func intsToStrings(nums []int) []string {
	s := make([]string, len(nums))
	for i, n := range nums {
		s[i] = fmt.Sprintf("%d", n)
	}
	return s
}

func sendDesktopNotification(title, body string) error {
	switch runtime.GOOS {
	case "darwin":
		script := fmt.Sprintf(`display notification %q with title %q`, body, title)
		return exec.Command("osascript", "-e", script).Run()
	case "linux":
		return exec.Command("notify-send", title, body).Run()
	case "windows":
		// PowerShell toast notification
		ps := fmt.Sprintf(
			`[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType=WindowsRuntime] | Out-Null; `+
				`$t = [Windows.UI.Notifications.ToastTemplateType]::ToastText02; `+
				`$x = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent($t); `+
				`$x.GetElementsByTagName('text')[0].AppendChild($x.CreateTextNode('%s')) | Out-Null; `+
				`$x.GetElementsByTagName('text')[1].AppendChild($x.CreateTextNode('%s')) | Out-Null; `+
				`[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('portwatch').Show($x)`,
			title, body,
		)
		return exec.Command("powershell", "-Command", ps).Run()
	default:
		return fmt.Errorf("desktop notifications not supported on %s", runtime.GOOS)
	}
}
