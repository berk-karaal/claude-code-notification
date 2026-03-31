package sound

import (
	"fmt"
	"runtime"

	"github.com/berk-karaal/claude-code-notification/internal/backend"
)

type Backend struct {
	OnNotification string
	OnStop         string
	OnStopFailure  string
}

func NewBackend(onStop, onNotification, onStopFailure string) *Backend {
	return &Backend{
		OnStop:         onStop,
		OnNotification: onNotification,
		OnStopFailure:  onStopFailure,
	}
}

func (b *Backend) resolveFile(event backend.EventType) (string, error) {
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
		return "", fmt.Errorf("sound backend: unsupported event type: %s", event)
	}
}

func defaultSoundFile() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return "/System/Library/Sounds/Glass.aiff", nil
	case "linux":
		if _, err := lookPath("paplay"); err == nil {
			return "/usr/share/sounds/freedesktop/stereo/bell.oga", nil
		}
		if _, err := lookPath("aplay"); err == nil {
			return "/usr/share/sounds/freedesktop/stereo/bell.wav", nil
		}
		return "", fmt.Errorf("sound backend: no audio player found for default sound (tried paplay, aplay)")
	default:
		return "", fmt.Errorf("sound backend: unsupported platform: %s", runtime.GOOS)
	}
}

// Send plays a sound file for the given event.
// Non-blocking: uses cmd.Start() without cmd.Wait() so the binary can exit
// immediately while the audio process continues.
func (b *Backend) Send(payload backend.Payload) error {
	filePath, err := b.resolveFile(payload.Event)
	if err != nil {
		return fmt.Errorf("sound backend: %w", err)
	}

	if filePath == "" {
		filePath, err = defaultSoundFile()
		if err != nil {
			return fmt.Errorf("sound backend: failed to get default sound file: %w", err)
		}
	}

	var playerPath string
	var playerArgs []string
	switch runtime.GOOS {
	case "darwin":
		playerPath = "afplay"
		playerArgs = []string{filePath}
	case "linux":
		if path, err := lookPath("paplay"); err == nil {
			playerPath = path
			playerArgs = []string{filePath}
		} else if path, err := lookPath("aplay"); err == nil {
			playerPath = path
			playerArgs = []string{filePath}
		} else {
			return fmt.Errorf("sound backend: no audio player found (tried paplay, aplay): %s", runtime.GOOS)
		}
	default:
		return fmt.Errorf("sound backend: unsupported platform: %s", runtime.GOOS)
	}

	cmd := execCommand(playerPath, playerArgs...)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("sound backend: failed to start playback: %w", err)
	}

	return nil
}
