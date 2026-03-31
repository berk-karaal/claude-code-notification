package integration

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// createMinimalWAV writes a valid WAV file with zero audio samples.
// afplay/paplay/aplay will accept it and return immediately.
func createMinimalWAV(t *testing.T, path string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create WAV file: %v", err)
	}
	defer f.Close()

	// RIFF header
	f.Write([]byte("RIFF"))
	binary.Write(f, binary.LittleEndian, uint32(36)) // file size - 8
	f.Write([]byte("WAVE"))
	// fmt subchunk
	f.Write([]byte("fmt "))
	binary.Write(f, binary.LittleEndian, uint32(16)) // subchunk size
	binary.Write(f, binary.LittleEndian, uint16(1))  // PCM format
	binary.Write(f, binary.LittleEndian, uint16(1))  // mono
	binary.Write(f, binary.LittleEndian, uint32(44100)) // sample rate
	binary.Write(f, binary.LittleEndian, uint32(44100)) // byte rate
	binary.Write(f, binary.LittleEndian, uint16(1))  // block align
	binary.Write(f, binary.LittleEndian, uint16(8))  // bits per sample
	// data subchunk
	f.Write([]byte("data"))
	binary.Write(f, binary.LittleEndian, uint32(0)) // 0 bytes of audio
}

// TestSound_StopEvent_PlaysConfiguredFile verifies the binary attempts to play
// the configured sound file via the platform audio player.
func TestSound_StopEvent_PlaysConfiguredFile(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("sound backend only supported on macOS and Linux")
	}

	cwd := t.TempDir()
	wavFile := filepath.Join(cwd, "alert.wav")
	createMinimalWAV(t, wavFile)

	configJSON := `{
		"system_notification": {"enabled": false},
		"sound": {"enabled": true, "on_stop": "` + wavFile + `"}
	}`
	if err := os.WriteFile(filepath.Join(cwd, ".claude-code-notification.json"), []byte(configJSON), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, stderr, code := runBinary(t, makeInput("Stop", cwd))
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, stderr)
	}
	// Sound uses non-blocking Start() — binary exits before playback finishes.
	// No "sound:" error in stderr means the player was found and started.
	if strings.Contains(stderr, "sound: failed to start") {
		t.Errorf("sound playback failed to start: %s", stderr)
	}
}

// TestSound_NotificationEvent_PlaysOnNotificationFile verifies the on_notification
// sound file is used for Notification events.
func TestSound_NotificationEvent_PlaysOnNotificationFile(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("sound backend only supported on macOS and Linux")
	}

	cwd := t.TempDir()
	wavFile := filepath.Join(cwd, "notify.wav")
	createMinimalWAV(t, wavFile)

	configJSON := `{
		"system_notification": {"enabled": false},
		"sound": {"enabled": true, "on_notification": "` + wavFile + `"}
	}`
	if err := os.WriteFile(filepath.Join(cwd, ".claude-code-notification.json"), []byte(configJSON), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, stderr, code := runBinary(t, makeInput("Notification", cwd))
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, stderr)
	}
	if strings.Contains(stderr, "sound: failed to start") {
		t.Errorf("sound playback failed to start: %s", stderr)
	}
}

// TestSound_DefaultFile_UsedWhenNotConfigured verifies the OS default sound is
// used when sound is enabled but no file is specified (D-10, D-11).
func TestSound_DefaultFile_UsedWhenNotConfigured(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("default sound file test only reliable on macOS")
	}

	cwd := setupProjectDir(t, `{
		"system_notification": {"enabled": false},
		"sound": {"enabled": true}
	}`)

	_, stderr, code := runBinary(t, makeInput("Stop", cwd))
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, stderr)
	}
	if strings.Contains(stderr, "sound: failed to start") {
		t.Errorf("sound playback failed to start: %s", stderr)
	}
}
