package sound

import (
	"os/exec"
	"runtime"
	"testing"

	"github.com/berk-karaal/claude-code-notification/internal/backend"
	"github.com/stretchr/testify/assert"
)

func TestResolveFile_StopFailureFallsBackToOnStop(t *testing.T) {
	b := NewBackend("/path/stop.wav", "", "")
	got, err := b.resolveFile(backend.EventStopFailure)
	assert.NoError(t, err, "expected no error resolving file")
	assert.Equal(t, "/path/stop.wav", got, "expected OnStop fallback")
}

func TestResolveFile_StopFailureUsesOwnPath(t *testing.T) {
	b := NewBackend("/path/stop.wav", "", "/path/failure.wav")
	got, err := b.resolveFile(backend.EventStopFailure)
	assert.NoError(t, err, "expected no error resolving file")
	assert.Equal(t, "/path/failure.wav", got, "expected OnStopFailure path")
}

func TestResolveFile_OnNotification(t *testing.T) {
	b := NewBackend("/path/stop.wav", "/path/notify.wav", "")
	got, err := b.resolveFile(backend.EventNotification)
	assert.NoError(t, err, "expected no error resolving file")
	assert.Equal(t, "/path/notify.wav", got, "expected OnNotification path")
}

func TestResolveFile_OnStop(t *testing.T) {
	b := NewBackend("/path/stop.wav", "", "")
	got, err := b.resolveFile(backend.EventStop)
	assert.NoError(t, err, "expected no error resolving file")
	assert.Equal(t, "/path/stop.wav", got, "expected OnStop path")
}

func TestSend_CallsCorrectPlayer(t *testing.T) {
	var gotName string
	var gotArgs []string
	execCommand = func(name string, arg ...string) *exec.Cmd {
		gotName = name
		gotArgs = arg
		return exec.Command("true")
	}
	defer func() { execCommand = exec.Command }()

	lookPath = func(file string) (string, error) {
		if file == "paplay" {
			return "/usr/bin/paplay", nil
		}
		return "", exec.ErrNotFound
	}
	defer func() { lookPath = exec.LookPath }()

	b := NewBackend("/path/sound.wav", "", "")
	_ = b.Send(backend.Payload{Event: backend.EventStop})

	if runtime.GOOS == "darwin" {
		assert.Equal(t, "afplay", gotName, "expected 'afplay' on darwin")
		assert.NotEmpty(t, gotArgs)
		assert.Equal(t, "/path/sound.wav", gotArgs[0])
	} else if runtime.GOOS == "linux" {
		assert.Equal(t, "/usr/bin/paplay", gotName, "expected '/usr/bin/paplay' on linux")
		assert.NotEmpty(t, gotArgs)
		assert.Equal(t, "/path/sound.wav", gotArgs[0])
	}
}

func TestSend_NoSoundFile_NoPlayerAvailable_ReturnsError(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("skipping: darwin always has afplay available via hardcoded path")
	}

	execCommand = func(name string, arg ...string) *exec.Cmd {
		t.Errorf("execCommand should not be called when no player is available")
		return exec.Command("true")
	}
	defer func() { execCommand = exec.Command }()

	lookPath = func(file string) (string, error) {
		return "", exec.ErrNotFound
	}
	defer func() { lookPath = exec.LookPath }()

	b := NewBackend("", "", "")
	err := b.Send(backend.Payload{Event: backend.EventStop})
	assert.Error(t, err, "expected error when no player available")
}

func TestSend_AllThreeEvents(t *testing.T) {
	cases := []struct {
		event    backend.EventType
		filePath string
	}{
		{backend.EventStop, "/path/stop.wav"},
		{backend.EventNotification, "/path/notify.wav"},
		{backend.EventStopFailure, "/path/failure.wav"},
	}

	for _, tc := range cases {
		t.Run(string(tc.event), func(t *testing.T) {
			var gotArgs []string
			execCommand = func(name string, arg ...string) *exec.Cmd {
				gotArgs = arg
				return exec.Command("true")
			}
			defer func() { execCommand = exec.Command }()

			lookPath = func(file string) (string, error) {
				if file == "paplay" {
					return "/usr/bin/paplay", nil
				}
				return "", exec.ErrNotFound
			}
			defer func() { lookPath = exec.LookPath }()

			b := NewBackend("/path/stop.wav", "/path/notify.wav", "/path/failure.wav")
			err := b.Send(backend.Payload{Event: tc.event})
			assert.NoError(t, err, "Send returned error for event %s", tc.event)
			assert.NotEmpty(t, gotArgs, "expected args for event %s", tc.event)
			assert.Equal(t, tc.filePath, gotArgs[0], "wrong file path for event %s", tc.event)
		})
	}
}
