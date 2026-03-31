package script

import (
	"fmt"
	"os"

	"github.com/berk-karaal/claude-code-notification/internal/backend"
)

type Backend struct {
	OnNotification string
	OnStop         string
	OnStopFailure  string
}

func NewBackend(onStop, onNotification, onStopFailure string) (*Backend, error) {
	if onStop == "" && onNotification == "" && onStopFailure == "" {
		return nil, fmt.Errorf("script backend: at least one script path must be configured")
	}

	return &Backend{
		OnStop:         onStop,
		OnNotification: onNotification,
		OnStopFailure:  onStopFailure,
	}, nil
}

func (b *Backend) resolveScript(event backend.EventType) (string, error) {
	switch event {
	case backend.EventNotification:
		return b.OnNotification, nil
	case backend.EventStopFailure:
		if b.OnStopFailure != "" {
			return b.OnStopFailure, nil
		}
		return b.OnStop, nil
	case backend.EventStop:
		return b.OnStop, nil
	default:
		return "", fmt.Errorf("script backend: unsupported event type: %s", event)
	}
}

func (b *Backend) Send(payload backend.Payload) error {
	scriptPath, err := b.resolveScript(payload.Event)
	if err != nil {
		return fmt.Errorf("script backend: failed to resolve script: %w", err)
	}

	if scriptPath == "" {
		return nil
	}

	info, err := os.Stat(scriptPath)
	if err != nil {
		return fmt.Errorf("script backend: file not found: %s: %w", scriptPath, err)
	}

	if info.Mode().Perm()&0111 == 0 {
		return fmt.Errorf("script backend: not executable: %s", scriptPath)
	}

	cmd := execCommand(scriptPath)
	cmd.Env = append(os.Environ(),
		"CCN_EVENT="+string(payload.Event),
		"CCN_HOSTNAME="+payload.Hostname,
		"CCN_PROJECT="+payload.Project,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("script backend: execution failed: %s: %w", scriptPath, err)
	}
	return nil
}
