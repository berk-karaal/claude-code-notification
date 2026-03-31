package notification

import (
	"fmt"
	"os"
	"testing"

	"github.com/berk-karaal/claude-code-notification/internal/backend"
)

func TestBuild_StopEvent(t *testing.T) {
	payload := Build("Stop", "/Users/user/myproject")

	if payload.Title != "Claude Finished" {
		t.Errorf("expected title 'Claude Finished', got %q", payload.Title)
	}
	if payload.Event != backend.EventStop {
		t.Errorf("expected event %q, got %q", backend.EventStop, payload.Event)
	}
}

func TestBuild_NotificationEvent(t *testing.T) {
	payload := Build("Notification", "/Users/user/myproject")

	if payload.Title != "Claude Needs Input" {
		t.Errorf("expected title 'Claude Needs Input', got %q", payload.Title)
	}
	if payload.Event != backend.EventNotification {
		t.Errorf("expected event %q, got %q", backend.EventNotification, payload.Event)
	}
}

func TestBuild_StopFailureEvent(t *testing.T) {
	payload := Build("StopFailure", "/Users/user/myproject")

	if payload.Title != "Claude Error" {
		t.Errorf("expected title 'Claude Error', got %q", payload.Title)
	}
	if payload.Event != backend.EventStopFailure {
		t.Errorf("expected event %q, got %q", backend.EventStopFailure, payload.Event)
	}
}

func TestBuild_UnknownEvent(t *testing.T) {
	payload := Build("Unknown", "/some/path")

	if payload.Title != "Claude Notification" {
		t.Errorf("expected title 'Claude Notification', got %q", payload.Title)
	}
}

func TestBuild_BodyContainsBasenameNotFullPath(t *testing.T) {
	payload := Build("Stop", "/Users/user/deeply/nested/myproject")

	hostname, _ := os.Hostname()
	expected := fmt.Sprintf("[%s] Project: myproject", hostname)

	if payload.Body != expected {
		t.Errorf("expected body %q, got %q", expected, payload.Body)
	}
}

func TestBuild_BodyContainsHostname(t *testing.T) {
	payload := Build("Stop", "/some/project")

	hostname, _ := os.Hostname()
	expectedPrefix := fmt.Sprintf("[%s]", hostname)

	if len(payload.Body) < len(expectedPrefix) || payload.Body[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("expected body to start with %q, got %q", expectedPrefix, payload.Body)
	}
}

func TestBuild_BodyFormat(t *testing.T) {
	payload := Build("Notification", "/home/dev/workspace")

	hostname, _ := os.Hostname()
	expected := fmt.Sprintf("[%s] Project: workspace", hostname)

	if payload.Body != expected {
		t.Errorf("expected body %q, got %q", expected, payload.Body)
	}
}
