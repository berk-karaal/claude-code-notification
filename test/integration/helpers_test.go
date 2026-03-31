package integration

import (
	"os"
	"path/filepath"
	"testing"
)

// setupProjectDir creates a temp directory with an optional project config file.
// Returns the directory path (used as cwd in hook event JSON).
func setupProjectDir(t *testing.T, configJSON string) string {
	t.Helper()
	cwd := t.TempDir()
	if configJSON != "" {
		err := os.WriteFile(filepath.Join(cwd, ".claude-code-notification.json"), []byte(configJSON), 0644)
		if err != nil {
			t.Fatalf("failed to write project config: %v", err)
		}
	}
	return cwd
}

// makeInput builds a hook event JSON string for the given event and cwd.
func makeInput(event, cwd string) string {
	return `{"hook_event_name":"` + event + `","cwd":"` + cwd + `","stop_hook_active":false}`
}
