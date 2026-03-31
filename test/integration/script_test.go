package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestScript_StopEvent_ExecutesScript verifies the binary runs the configured
// custom script and passes CCN_ environment variables.
func TestScript_StopEvent_ExecutesScript(t *testing.T) {
	cwd := t.TempDir()
	markerFile := filepath.Join(cwd, "marker.txt")

	// Create a script that writes env vars to a marker file
	scriptPath := filepath.Join(cwd, "on_stop.sh")
	scriptContent := "#!/bin/sh\necho \"event=$CCN_EVENT hostname=$CCN_HOSTNAME project=$CCN_PROJECT\" > " + markerFile + "\n"
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	configJSON := `{
		"system_notification": {"enabled": false},
		"custom_script": {"enabled": true, "on_stop": "` + scriptPath + `"}
	}`
	if err := os.WriteFile(filepath.Join(cwd, ".claude-code-notification.json"), []byte(configJSON), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, stderr, code := runBinary(t, makeInput("Stop", cwd))
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, stderr)
	}

	data, err := os.ReadFile(markerFile)
	if err != nil {
		t.Fatalf("marker file not created — script did not run: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "event=Stop") {
		t.Errorf("expected CCN_EVENT=Stop in marker, got %q", output)
	}
	if !strings.Contains(output, "project="+filepath.Base(cwd)) {
		t.Errorf("expected CCN_PROJECT=%s in marker, got %q", filepath.Base(cwd), output)
	}
}

// TestScript_NotificationEvent_ExecutesOnNotification verifies the on_notification
// script is called for Notification events.
func TestScript_NotificationEvent_ExecutesOnNotification(t *testing.T) {
	cwd := t.TempDir()
	markerFile := filepath.Join(cwd, "marker.txt")

	scriptPath := filepath.Join(cwd, "on_notification.sh")
	scriptContent := "#!/bin/sh\necho \"$CCN_EVENT\" > " + markerFile + "\n"
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	configJSON := `{
		"system_notification": {"enabled": false},
		"custom_script": {"enabled": true, "on_notification": "` + scriptPath + `"}
	}`
	if err := os.WriteFile(filepath.Join(cwd, ".claude-code-notification.json"), []byte(configJSON), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, stderr, code := runBinary(t, makeInput("Notification", cwd))
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, stderr)
	}

	data, err := os.ReadFile(markerFile)
	if err != nil {
		t.Fatalf("marker file not created — script did not run: %v", err)
	}
	if got := strings.TrimSpace(string(data)); got != "Notification" {
		t.Errorf("expected 'Notification', got %q", got)
	}
}

// TestScript_StopFailureEvent_FallsBackToOnStop verifies that StopFailure
// falls back to on_stop when on_stop_failure is not configured (D-02).
func TestScript_StopFailureEvent_FallsBackToOnStop(t *testing.T) {
	cwd := t.TempDir()
	markerFile := filepath.Join(cwd, "marker.txt")

	scriptPath := filepath.Join(cwd, "on_stop.sh")
	scriptContent := "#!/bin/sh\necho \"$CCN_EVENT\" > " + markerFile + "\n"
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	configJSON := `{
		"system_notification": {"enabled": false},
		"custom_script": {"enabled": true, "on_stop": "` + scriptPath + `"}
	}`
	if err := os.WriteFile(filepath.Join(cwd, ".claude-code-notification.json"), []byte(configJSON), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, stderr, code := runBinary(t, makeInput("StopFailure", cwd))
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, stderr)
	}

	data, err := os.ReadFile(markerFile)
	if err != nil {
		t.Fatalf("marker file not created — StopFailure did not fall back to on_stop: %v", err)
	}
	if got := strings.TrimSpace(string(data)); got != "StopFailure" {
		t.Errorf("expected 'StopFailure', got %q", got)
	}
}
