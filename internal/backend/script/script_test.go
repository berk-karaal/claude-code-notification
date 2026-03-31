package script

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/berk-karaal/claude-code-notification/internal/backend"
	"github.com/stretchr/testify/assert"
)

func TestNewBackend_AllEmpty_ReturnsError(t *testing.T) {
	_, err := NewBackend("", "", "")
	assert.Error(t, err, "expected error when all script paths are empty")
}

func TestSend_ExecutesConfiguredScript(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test-script.sh")
	err := os.WriteFile(scriptPath, []byte("#!/bin/sh\n"), 0755)
	assert.NoError(t, err, "failed to create script")

	var gotName string
	execCommand = func(name string, arg ...string) *exec.Cmd {
		gotName = name
		return exec.Command("true")
	}
	defer func() { execCommand = exec.Command }()

	b, err := NewBackend(scriptPath, "", "")
	assert.NoError(t, err, "failed to create backend")

	err = b.Send(backend.Payload{Event: backend.EventStop})
	assert.NoError(t, err, "Send returned unexpected error")
	assert.Equal(t, scriptPath, gotName, "expected execCommand called with script path")
}

func TestSend_PassesCCNEnvVars(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test-script.sh")
	err := os.WriteFile(scriptPath, []byte("#!/bin/sh\n"), 0755)
	assert.NoError(t, err, "failed to create script")

	var capturedCmd *exec.Cmd
	execCommand = func(name string, arg ...string) *exec.Cmd {
		cmd := exec.Command("true")
		capturedCmd = cmd
		return cmd
	}
	defer func() { execCommand = exec.Command }()

	b, err := NewBackend(scriptPath, "", "")
	assert.NoError(t, err, "failed to create backend")

	payload := backend.Payload{
		Event:    backend.EventStop,
		Hostname: "myhost",
		Project:  "myproject",
	}
	_ = b.Send(payload)

	assert.NotNil(t, capturedCmd, "execCommand was not called")

	envMap := make(map[string]string)
	for _, e := range capturedCmd.Env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	assert.Equal(t, "Stop", envMap["CCN_EVENT"])
	assert.Equal(t, "myhost", envMap["CCN_HOSTNAME"])
	assert.Equal(t, "myproject", envMap["CCN_PROJECT"])
}

func TestSend_MissingFile_ReturnsError(t *testing.T) {
	var called bool
	execCommand = func(name string, arg ...string) *exec.Cmd {
		called = true
		return exec.Command("true")
	}
	defer func() { execCommand = exec.Command }()

	b, err := NewBackend("/nonexistent/path/script.sh", "", "")
	assert.NoError(t, err, "failed to create backend")

	err = b.Send(backend.Payload{Event: backend.EventStop})
	assert.Error(t, err, "expected error for missing file")
	assert.ErrorContains(t, err, "script backend: file not found:")
	assert.False(t, called, "expected execCommand NOT to be called for missing file")
}

func TestSend_NonExecutableFile_ReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "no-exec.sh")
	err := os.WriteFile(scriptPath, []byte("#!/bin/sh\n"), 0644)
	assert.NoError(t, err, "failed to create script")

	var called bool
	execCommand = func(name string, arg ...string) *exec.Cmd {
		called = true
		return exec.Command("true")
	}
	defer func() { execCommand = exec.Command }()

	b, err := NewBackend(scriptPath, "", "")
	assert.NoError(t, err, "failed to create backend")

	err = b.Send(backend.Payload{Event: backend.EventStop})
	assert.Error(t, err, "expected error for non-executable file")
	assert.ErrorContains(t, err, "script backend: not executable:")
	assert.False(t, called, "expected execCommand NOT to be called for non-executable file")
}

func TestResolveScript_StopFailureFallback(t *testing.T) {
	b, err := NewBackend("/path/stop.sh", "", "")
	assert.NoError(t, err, "failed to create backend")

	got, err := b.resolveScript(backend.EventStopFailure)
	assert.NoError(t, err, "failed to resolve script")
	assert.Equal(t, "/path/stop.sh", got, "expected OnStop fallback")
}

func TestSend_EmptyScriptPath_ReturnsError(t *testing.T) {
	var called bool
	execCommand = func(name string, arg ...string) *exec.Cmd {
		called = true
		return exec.Command("true")
	}
	defer func() { execCommand = exec.Command }()

	b, err := NewBackend("/path/stop.sh", "", "")
	assert.NoError(t, err, "failed to create backend")

	// Notification event has no script configured
	err = b.Send(backend.Payload{Event: backend.EventNotification})
	assert.NoError(t, err, "expected nil error for empty script path")
	assert.False(t, called, "expected execCommand NOT to be called for empty script path")
}

func TestSend_AllThreeEvents(t *testing.T) {
	tmpDir := t.TempDir()

	stopScript := filepath.Join(tmpDir, "stop.sh")
	notifyScript := filepath.Join(tmpDir, "notify.sh")
	failureScript := filepath.Join(tmpDir, "failure.sh")
	for _, path := range []string{stopScript, notifyScript, failureScript} {
		err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0755)
		assert.NoError(t, err, "failed to create script %s", path)
	}

	cases := []struct {
		event          backend.EventType
		expectedScript string
	}{
		{backend.EventStop, stopScript},
		{backend.EventNotification, notifyScript},
		{backend.EventStopFailure, failureScript},
	}

	for _, tc := range cases {
		t.Run(string(tc.event), func(t *testing.T) {
			var gotName string
			execCommand = func(name string, arg ...string) *exec.Cmd {
				gotName = name
				return exec.Command("true")
			}
			defer func() { execCommand = exec.Command }()

			b, err := NewBackend(stopScript, notifyScript, failureScript)
			assert.NoError(t, err, "failed to create backend")

			err = b.Send(backend.Payload{Event: tc.event, Hostname: "host", Project: "proj"})
			assert.NoError(t, err, "Send returned error for event %s", tc.event)
			assert.Equal(t, tc.expectedScript, gotName, "wrong script for event %s", tc.event)
		})
	}
}

func Test_UnknownEvent(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test-script.sh")
	err := os.WriteFile(scriptPath, []byte("#!/bin/sh\n"), 0755)
	assert.NoError(t, err, "failed to create script")

	var gotName string
	execCommand = func(name string, arg ...string) *exec.Cmd {
		gotName = name
		return exec.Command("true")
	}
	defer func() { execCommand = exec.Command }()

	b, err := NewBackend(scriptPath, "", "")
	assert.NoError(t, err, "failed to create backend")

	err = b.Send(backend.Payload{Event: backend.EventType("test-event")})
	assert.Error(t, err, "expected error for unknown event type")
	assert.ErrorContains(t, err, "script backend: unsupported event type")
	assert.False(t, gotName != "", "expected execCommand NOT to be called for unknown event type")
}

func Test_ExecFailure_ReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test-script.sh")
	err := os.WriteFile(scriptPath, []byte("#!/bin/sh\nexit 1\n"), 0755)
	assert.NoError(t, err, "failed to create script")

	b, err := NewBackend(scriptPath, "", "")
	assert.NoError(t, err, "failed to create backend")

	err = b.Send(backend.Payload{Event: backend.EventStop})
	assert.Error(t, err, "expected error when script exits with non-zero status")
	assert.ErrorContains(t, err, "script backend: execution failed:")
}
