package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/berk-karaal/claude-code-notification/internal/backend"
	"github.com/berk-karaal/claude-code-notification/internal/backend/ntfy"
	"github.com/berk-karaal/claude-code-notification/internal/backend/script"
	"github.com/berk-karaal/claude-code-notification/internal/backend/sound"
	"github.com/berk-karaal/claude-code-notification/internal/backend/system"
	"github.com/berk-karaal/claude-code-notification/internal/config"
	"github.com/berk-karaal/claude-code-notification/internal/notification"
)

// version is set at build time via: go build -ldflags "-X main.version=0.1.0"
// The bash wrapper checks this via: ${BINARY} --version
var version = "###"

// hookEvent holds the fields consumed from Claude Code hook stdin JSON. Only hook_event_name, cwd,
// and stop_hook_active are used. All other fields are intentionally ignored — they shouldn't be
// used in the notification content for security and privacy reasons.
type hookEvent struct {
	HookEventName  string `json:"hook_event_name"`
	Cwd            string `json:"cwd"`
	StopHookActive bool   `json:"stop_hook_active"`
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println(version)
		return
	}

	var event hookEvent
	dec := json.NewDecoder(os.Stdin)
	if err := dec.Decode(&event); err != nil {
		log.Printf("failed to decode hook event from stdin: %v", err)
		return
	}

	// Guard against Stop hook infinite loop.
	// If stop_hook_active is true, a hook is already running — exit immediately.
	if event.HookEventName == "Stop" && event.StopHookActive {
		return
	}

	cfg, err := config.Load(event.Cwd)
	if err != nil {
		log.Printf("config load error (using defaults): %v", err)
		cfg = config.DefaultConfig()
	}

	payload := notification.Build(event.HookEventName, event.Cwd)

	enabledBackends, errs := getEnabledBackends(cfg)
	if len(errs) > 0 {
		for _, err := range errs {
			log.Printf("backend init error: %v", err)
		}
		return
	}

	for _, b := range enabledBackends {
		if err := b.Send(payload); err != nil {
			log.Printf("backend send error: %v", err)
		}
	}
}

// getEnabledBackends returns the list of backends that are enabled in the user configuration.
func getEnabledBackends(cfg config.Config) ([]backend.Backend, []error) {
	var backends []backend.Backend
	errors := []error{}

	if cfg.SystemNotification.Enabled {
		backends = append(backends, system.NewBackend())
	}

	if cfg.Ntfy.Enabled {
		b, err := ntfy.NewBackend(cfg.Ntfy.Topic, cfg.Ntfy.Server)
		if err != nil {
			errors = append(errors, fmt.Errorf("ntfy backend init error: %w", err))
		} else {
			backends = append(backends, b)
		}
	}

	if cfg.Sound.Enabled {
		b := sound.NewBackend(cfg.Sound.OnStop, cfg.Sound.OnNotification, cfg.Sound.OnStopFailure)
		backends = append(backends, b)
	}

	if cfg.CustomScript.Enabled {
		b, err := script.NewBackend(cfg.CustomScript.OnStop, cfg.CustomScript.OnNotification, cfg.CustomScript.OnStopFailure)
		if err != nil {
			errors = append(errors, fmt.Errorf("custom script backend init error: %w", err))
		} else {
			backends = append(backends, b)
		}
	}

	return backends, errors
}
