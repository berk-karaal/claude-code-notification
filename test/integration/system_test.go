package integration

import (
	"runtime"
	"strings"
	"testing"
)

// TestSystem_StopEvent_SendsNotification verifies the system notification backend
// dispatches via osascript (macOS) or notify-send (Linux) for a Stop event.
func TestSystem_StopEvent_SendsNotification(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("system notification only supported on macOS and Linux")
	}

	// system_notification is enabled by default — no explicit config needed
	cwd := setupProjectDir(t, `{"system_notification": {"enabled": true}}`)

	_, stderr, code := runBinary(t, makeInput("Stop", cwd))
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, stderr)
	}
	// On macOS, osascript should succeed silently. On Linux without notify-send,
	// "not found" is logged but exit is still 0.
	if runtime.GOOS == "darwin" && strings.Contains(stderr, "osascript failed") {
		t.Errorf("osascript failed unexpectedly: %s", stderr)
	}
}

// TestSystem_NotificationEvent_SendsNotification verifies the Notification event
// dispatches a system notification.
func TestSystem_NotificationEvent_SendsNotification(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("system notification only supported on macOS and Linux")
	}

	cwd := setupProjectDir(t, `{"system_notification": {"enabled": true}}`)

	_, stderr, code := runBinary(t, makeInput("Notification", cwd))
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, stderr)
	}
	if runtime.GOOS == "darwin" && strings.Contains(stderr, "osascript failed") {
		t.Errorf("osascript failed unexpectedly: %s", stderr)
	}
}

// TestSystem_StopFailureEvent_SendsNotification verifies the StopFailure event
// dispatches a system notification.
func TestSystem_StopFailureEvent_SendsNotification(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("system notification only supported on macOS and Linux")
	}

	cwd := setupProjectDir(t, `{"system_notification": {"enabled": true}}`)

	_, stderr, code := runBinary(t, makeInput("StopFailure", cwd))
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, stderr)
	}
	if runtime.GOOS == "darwin" && strings.Contains(stderr, "osascript failed") {
		t.Errorf("osascript failed unexpectedly: %s", stderr)
	}
}
