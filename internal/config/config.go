package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Config is the top-level configuration structure.
// Each backend section has an Enabled field and backend-specific settings.
// Matches the PRD config schema exactly (D-07).
type Config struct {
	Ntfy               NtfyConfig              `json:"ntfy"`
	SystemNotification SystemNotificationConfig `json:"system_notification"`
	Sound              SoundConfig             `json:"sound"`
	CustomScript       CustomScriptConfig      `json:"custom_script"`
}

// NtfyConfig holds settings for the ntfy push notification backend.
type NtfyConfig struct {
	Enabled bool   `json:"enabled"`
	Topic   string `json:"topic"`
	Server  string `json:"server"`
}

// SystemNotificationConfig holds settings for the OS system notification backend.
type SystemNotificationConfig struct {
	Enabled bool `json:"enabled"`
}

// SoundConfig holds settings for the sound playback backend.
type SoundConfig struct {
	Enabled        bool   `json:"enabled"`
	OnNotification string `json:"on_notification"`
	OnStop         string `json:"on_stop"`
	OnStopFailure  string `json:"on_stop_failure"`
}

// CustomScriptConfig holds settings for the custom script backend.
type CustomScriptConfig struct {
	Enabled        bool   `json:"enabled"`
	OnNotification string `json:"on_notification"`
	OnStop         string `json:"on_stop"`
	OnStopFailure  string `json:"on_stop_failure"`
}

// DefaultConfig returns safe defaults used when no config file is found.
// D-08: SystemNotification is enabled by default so zero-config installs
// still dispatch desktop notifications.
func DefaultConfig() Config {
	return Config{
		SystemNotification: SystemNotificationConfig{Enabled: true},
	}
}

// Load returns the effective config for the current invocation.
// Resolution order (3-layer merge chain):
//  1. DefaultConfig() — base layer
//  2. Global ~/.config/claude-code-notification/config.json — merges over defaults
//  3. Per-project .claude-code-notification.json in cwd — merges over global
//
// Fields not present in a higher-priority file are inherited from the layer below.
// Setting enabled:false in project config overrides a globally-enabled backend.
func Load(cwd string) (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		// Cannot determine home dir; skip global config, use defaults + project
		home = ""
	}
	return loadWithHome(cwd, home)
}

// loadWithHome is the testable core of Load, accepting the home directory as a
// parameter so tests can inject a temporary directory instead of the real home.
func loadWithHome(cwd, homeDir string) (Config, error) {
	cfg := DefaultConfig()

	// Layer 1: global config merges over defaults
	if homeDir != "" {
		globalPath := filepath.Join(homeDir, ".config", "claude-code-notification", "config.json")
		if err := mergeFile(&cfg, globalPath); err != nil {
			return cfg, err
		}
	}

	// Layer 2: project config merges over global
	if cwd != "" {
		projectPath := filepath.Join(cwd, ".claude-code-notification.json")
		if err := mergeFile(&cfg, projectPath); err != nil {
			return cfg, err
		}
	}

	return cfg, nil
}

// mergeFile decodes a JSON config file at path into the existing cfg struct.
// Only fields present in the JSON are overwritten; other fields remain unchanged.
// If the file does not exist (or cannot be opened), mergeFile returns nil silently.
// If the file exists but contains invalid JSON, the parse error is returned.
func mergeFile(cfg *Config, path string) error {
	f, err := os.Open(path)
	if err != nil {
		// File not found or unreadable — silently skip
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		// Other OS errors (permission denied, etc.) also skip silently
		return nil
	}
	defer f.Close()

	// Decode into the existing struct — only present JSON fields are overwritten
	if err := json.NewDecoder(f).Decode(cfg); err != nil {
		return err
	}
	return nil
}
