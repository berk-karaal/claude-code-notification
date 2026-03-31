package system

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/berk-karaal/claude-code-notification/internal/backend"
)

var asEscaper = strings.NewReplacer(`\`, `\\`, `"`, `\"`)

type Backend struct{}

func NewBackend() *Backend {
	return &Backend{}
}

func (b *Backend) Send(payload backend.Payload) error {
	switch runtime.GOOS {
	case "darwin":
		return b.sendDarwin(payload)
	case "linux":
		return b.sendLinux(payload)
	default:
		return fmt.Errorf("system backend: unsupported platform: %s", runtime.GOOS)
	}
}

func (b *Backend) sendDarwin(payload backend.Payload) error {
	safeTitle := asEscaper.Replace(payload.Title)
	safeBody := asEscaper.Replace(payload.Body)
	script := fmt.Sprintf(`display notification "%s" with title "%s"`, safeBody, safeTitle)

	cmd := execCommand("osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("system backend: osascript failed: %w", err)
	}

	return nil
}

func (b *Backend) sendLinux(payload backend.Payload) error {
	if _, err := lookPath("notify-send"); err != nil {
		return fmt.Errorf("system backend: notify-send not found")
	}

	cmd := execCommand("notify-send", payload.Title, payload.Body)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("system backend: notify-send failed: %w", err)
	}

	return nil
}
