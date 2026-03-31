package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig_SystemNotificationEnabled(t *testing.T) {
	cfg := DefaultConfig()
	if !cfg.SystemNotification.Enabled {
		t.Error("DefaultConfig().SystemNotification.Enabled should be true")
	}
}

func TestDefaultConfig_OtherBackendsDisabled(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Ntfy.Enabled {
		t.Error("DefaultConfig().Ntfy.Enabled should be false")
	}
	if cfg.Sound.Enabled {
		t.Error("DefaultConfig().Sound.Enabled should be false")
	}
	if cfg.CustomScript.Enabled {
		t.Error("DefaultConfig().CustomScript.Enabled should be false")
	}
}

func TestLoad_NoConfigReturnsDefault(t *testing.T) {
	// Load with empty cwd and no global config should return DefaultConfig.
	// Use loadWithHome with a temp dir to avoid reading the developer's real
	// ~/.config/claude-code-notification/config.json.
	cfg, err := loadWithHome("", t.TempDir())
	if err != nil {
		t.Fatalf("loadWithHome returned error: %v", err)
	}
	if !cfg.SystemNotification.Enabled {
		t.Error("Load with no configs should return default with SystemNotification.Enabled=true")
	}
	if cfg.Ntfy.Enabled {
		t.Error("Load with no configs should return default with Ntfy.Enabled=false")
	}
}

func TestLoad_ProjectConfigMergesWithDefaults(t *testing.T) {
	// Project config sets ntfy.enabled=true but does NOT set system_notification.
	// With merge semantics, system_notification.enabled should still be true from defaults.
	tmpDir := t.TempDir()
	tmpHome := t.TempDir()

	projectConfig := `{"ntfy":{"enabled":true,"topic":"test","server":"https://ntfy.sh"}}`
	err := os.WriteFile(filepath.Join(tmpDir, ".claude-code-notification.json"), []byte(projectConfig), 0644)
	if err != nil {
		t.Fatalf("failed to write project config: %v", err)
	}

	cfg, err := loadWithHome(tmpDir, tmpHome)
	if err != nil {
		t.Fatalf("loadWithHome returned error: %v", err)
	}

	if !cfg.Ntfy.Enabled {
		t.Error("project config has ntfy.enabled=true, but result has false")
	}
	// Merge semantics: system_notification.enabled comes from defaults
	if !cfg.SystemNotification.Enabled {
		t.Error("system_notification.enabled should be true from defaults (merge, not replace)")
	}
}

func TestLoad_GlobalConfigUsedWhenNoProjectConfig(t *testing.T) {
	// Create a temp project dir with no .claude-code-notification.json,
	// and a temp home dir with no global config.
	// Use loadWithHome to avoid reading the developer's real home directory.
	projectDir := t.TempDir()
	cfg, err := loadWithHome(projectDir, t.TempDir())
	if err != nil {
		t.Fatalf("loadWithHome returned error: %v", err)
	}

	// Without any config files, we should get defaults
	if !cfg.SystemNotification.Enabled {
		t.Error("expected SystemNotification.Enabled=true from defaults")
	}
}

func TestLoad_ProjectConfigInheritsDefaults(t *testing.T) {
	// Project config sets only sound.enabled=true.
	// Merge semantics: system_notification.enabled=true is preserved from defaults.
	tmpDir := t.TempDir()
	tmpHome := t.TempDir()

	projectConfig := `{"sound":{"enabled":true}}`
	err := os.WriteFile(filepath.Join(tmpDir, ".claude-code-notification.json"), []byte(projectConfig), 0644)
	if err != nil {
		t.Fatalf("failed to write project config: %v", err)
	}

	cfg, err := loadWithHome(tmpDir, tmpHome)
	if err != nil {
		t.Fatalf("loadWithHome returned error: %v", err)
	}

	if !cfg.Sound.Enabled {
		t.Error("sound should be enabled from project config")
	}
	// Merge semantics: system_notification.enabled=true from defaults is NOT zeroed
	if !cfg.SystemNotification.Enabled {
		t.Error("system_notification.enabled should be true (inherited from defaults, not zeroed by project config)")
	}
	if cfg.Ntfy.Enabled {
		t.Error("ntfy should be false (not set in any config)")
	}
	if cfg.CustomScript.Enabled {
		t.Error("custom_script should be false (not set in any config)")
	}
}

// TestLoad_MergeChain_DefaultsGlobalProject verifies the full 3-layer merge chain.
// Layer 1: defaults (system_notification.enabled=true)
// Layer 2: global config (ntfy.topic="global-topic", ntfy.server="https://custom.sh")
// Layer 3: project config (ntfy.enabled=true)
// Expected: ntfy.enabled=true, ntfy.topic="global-topic", ntfy.server="https://custom.sh", system_notification.enabled=true
func TestLoad_MergeChain_DefaultsGlobalProject(t *testing.T) {
	tmpHome := t.TempDir()
	tmpProject := t.TempDir()

	// Write global config
	globalDir := filepath.Join(tmpHome, ".config", "claude-code-notification")
	if err := os.MkdirAll(globalDir, 0755); err != nil {
		t.Fatalf("failed to create global config dir: %v", err)
	}
	globalConfig := `{"ntfy":{"topic":"global-topic","server":"https://custom.sh"}}`
	if err := os.WriteFile(filepath.Join(globalDir, "config.json"), []byte(globalConfig), 0644); err != nil {
		t.Fatalf("failed to write global config: %v", err)
	}

	// Write project config
	projectConfig := `{"ntfy":{"enabled":true}}`
	if err := os.WriteFile(filepath.Join(tmpProject, ".claude-code-notification.json"), []byte(projectConfig), 0644); err != nil {
		t.Fatalf("failed to write project config: %v", err)
	}

	cfg, err := loadWithHome(tmpProject, tmpHome)
	if err != nil {
		t.Fatalf("loadWithHome returned error: %v", err)
	}

	if !cfg.Ntfy.Enabled {
		t.Error("ntfy.enabled should be true from project config")
	}
	if cfg.Ntfy.Topic != "global-topic" {
		t.Errorf("ntfy.topic should be 'global-topic' from global config, got %q", cfg.Ntfy.Topic)
	}
	if cfg.Ntfy.Server != "https://custom.sh" {
		t.Errorf("ntfy.server should be 'https://custom.sh' from global config, got %q", cfg.Ntfy.Server)
	}
	if !cfg.SystemNotification.Enabled {
		t.Error("system_notification.enabled should be true from defaults")
	}
}

// TestLoad_ProjectExplicitDisable verifies that setting enabled:false in project config
// overrides a globally-enabled backend.
func TestLoad_ProjectExplicitDisable(t *testing.T) {
	tmpHome := t.TempDir()
	tmpProject := t.TempDir()

	// Global config enables system_notification explicitly
	globalDir := filepath.Join(tmpHome, ".config", "claude-code-notification")
	if err := os.MkdirAll(globalDir, 0755); err != nil {
		t.Fatalf("failed to create global config dir: %v", err)
	}
	globalConfig := `{"system_notification":{"enabled":true},"ntfy":{"enabled":true,"topic":"my-topic"}}`
	if err := os.WriteFile(filepath.Join(globalDir, "config.json"), []byte(globalConfig), 0644); err != nil {
		t.Fatalf("failed to write global config: %v", err)
	}

	// Project config explicitly disables system_notification
	projectConfig := `{"system_notification":{"enabled":false}}`
	if err := os.WriteFile(filepath.Join(tmpProject, ".claude-code-notification.json"), []byte(projectConfig), 0644); err != nil {
		t.Fatalf("failed to write project config: %v", err)
	}

	cfg, err := loadWithHome(tmpProject, tmpHome)
	if err != nil {
		t.Fatalf("loadWithHome returned error: %v", err)
	}

	if cfg.SystemNotification.Enabled {
		t.Error("system_notification.enabled should be false — project config explicitly set it to false")
	}
	// ntfy is inherited from global
	if !cfg.Ntfy.Enabled {
		t.Error("ntfy.enabled should be true (inherited from global config)")
	}
	if cfg.Ntfy.Topic != "my-topic" {
		t.Errorf("ntfy.topic should be 'my-topic' from global config, got %q", cfg.Ntfy.Topic)
	}
}

// TestLoad_ProjectOmittedFieldInheritsGlobal verifies that when project config
// has no sound section, the global sound settings are preserved.
func TestLoad_ProjectOmittedFieldInheritsGlobal(t *testing.T) {
	tmpHome := t.TempDir()
	tmpProject := t.TempDir()

	// Global config sets sound settings
	globalDir := filepath.Join(tmpHome, ".config", "claude-code-notification")
	if err := os.MkdirAll(globalDir, 0755); err != nil {
		t.Fatalf("failed to create global config dir: %v", err)
	}
	globalConfig := `{"sound":{"on_stop":"/path/to/sound.wav"}}`
	if err := os.WriteFile(filepath.Join(globalDir, "config.json"), []byte(globalConfig), 0644); err != nil {
		t.Fatalf("failed to write global config: %v", err)
	}

	// Project config has no sound section
	projectConfig := `{"ntfy":{"enabled":true}}`
	if err := os.WriteFile(filepath.Join(tmpProject, ".claude-code-notification.json"), []byte(projectConfig), 0644); err != nil {
		t.Fatalf("failed to write project config: %v", err)
	}

	cfg, err := loadWithHome(tmpProject, tmpHome)
	if err != nil {
		t.Fatalf("loadWithHome returned error: %v", err)
	}

	if cfg.Sound.OnStop != "/path/to/sound.wav" {
		t.Errorf("sound.on_stop should be '/path/to/sound.wav' from global config, got %q", cfg.Sound.OnStop)
	}
}

func TestMergeFile_NonexistentPathReturnsNil(t *testing.T) {
	// mergeFile on nonexistent path should return nil (silently skip missing files)
	cfg := DefaultConfig()
	err := mergeFile(&cfg, "/nonexistent/path/config.json")
	if err != nil {
		t.Errorf("mergeFile on nonexistent path should return nil, got: %v", err)
	}
}

func TestMergeFile_ValidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configJSON := `{"ntfy":{"enabled":true,"topic":"test-topic","server":"https://ntfy.sh"},"system_notification":{"enabled":false}}`
	err := os.WriteFile(configPath, []byte(configJSON), 0644)
	if err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := DefaultConfig()
	err = mergeFile(&cfg, configPath)
	if err != nil {
		t.Fatalf("mergeFile returned error: %v", err)
	}

	if !cfg.Ntfy.Enabled {
		t.Error("expected ntfy.enabled=true")
	}
	if cfg.Ntfy.Topic != "test-topic" {
		t.Errorf("expected topic 'test-topic', got %q", cfg.Ntfy.Topic)
	}
	if cfg.Ntfy.Server != "https://ntfy.sh" {
		t.Errorf("expected server 'https://ntfy.sh', got %q", cfg.Ntfy.Server)
	}
	if cfg.SystemNotification.Enabled {
		t.Error("expected system_notification.enabled=false")
	}
}

func TestMergeFile_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	err := os.WriteFile(configPath, []byte("{invalid json}"), 0644)
	if err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := DefaultConfig()
	err = mergeFile(&cfg, configPath)
	if err == nil {
		t.Error("mergeFile on invalid JSON should return error")
	}
}
