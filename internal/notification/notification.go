package notification

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/berk-karaal/claude-code-notification/internal/backend"
)

// Build constructs a safe notification Payload from hook event data.
// Only eventName and cwd are used — no message text, session IDs, paths,
// tool I/O, or environment variables ever appear in the output (CONT-03).
func Build(eventName string, cwd string) backend.Payload {
	hostname, _ := os.Hostname()
	project := filepath.Base(cwd)

	var title string
	switch eventName {
	case "Stop":
		title = "Claude Finished"
	case "Notification":
		title = "Claude Needs Input"
	case "StopFailure":
		title = "Claude Error"
	default:
		title = "Claude Notification"
	}

	body := fmt.Sprintf("[%s] Project: %s", hostname, project)

	return backend.Payload{
		Title:    title,
		Body:     body,
		Event:    backend.EventType(eventName),
		Hostname: hostname,
		Project:  project,
	}
}
