package system

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/berk-karaal/claude-code-notification/internal/backend"
	"github.com/stretchr/testify/assert"
)

func TestSendDarwin_CallsOsascript(t *testing.T) {
	var gotName string
	var gotArgs []string
	execCommand = func(name string, arg ...string) *exec.Cmd {
		gotName = name
		gotArgs = arg
		return exec.Command("true")
	}
	defer func() { execCommand = exec.Command }()

	b := NewBackend()
	err := b.sendDarwin(backend.Payload{Title: "Test Title", Body: "Test Body", Event: backend.EventStop})
	assert.NoError(t, err)
	assert.Equal(t, "osascript", gotName)
	assert.True(t, len(gotArgs) >= 2 && gotArgs[0] == "-e", "expected args starting with '-e', got %v", gotArgs)
}

func TestSendDarwin_ScriptContainsTitleAndBody(t *testing.T) {
	var gotScript string
	execCommand = func(name string, arg ...string) *exec.Cmd {
		if len(arg) >= 2 {
			gotScript = arg[1]
		}
		return exec.Command("true")
	}
	defer func() { execCommand = exec.Command }()

	b := NewBackend()
	err := b.sendDarwin(backend.Payload{Title: "My Title", Body: "My Body", Event: backend.EventStop})
	assert.NoError(t, err)
	assert.Contains(t, gotScript, "display notification")
	assert.Contains(t, gotScript, "My Title")
	assert.Contains(t, gotScript, "My Body")
}

func TestSendDarwin_AppleScriptEscaping(t *testing.T) {
	var gotScript string
	execCommand = func(name string, arg ...string) *exec.Cmd {
		if len(arg) >= 2 {
			gotScript = arg[1]
		}
		return exec.Command("true")
	}
	defer func() { execCommand = exec.Command }()

	b := NewBackend()
	err := b.sendDarwin(backend.Payload{
		Title: `He said "hello"`,
		Body:  `path\to\file`,
		Event: backend.EventStop,
	})
	assert.NoError(t, err)
	assert.Contains(t, gotScript, `He said \"hello\"`)
	assert.Contains(t, gotScript, `path\\to\\file`)
}

func TestSendLinux_CallsNotifySend(t *testing.T) {
	var gotName string
	var gotArgs []string
	execCommand = func(name string, arg ...string) *exec.Cmd {
		gotName = name
		gotArgs = arg
		return exec.Command("true")
	}
	defer func() { execCommand = exec.Command }()

	lookPath = func(file string) (string, error) {
		return "/usr/bin/notify-send", nil
	}
	defer func() { lookPath = exec.LookPath }()

	b := NewBackend()
	err := b.sendLinux(backend.Payload{Title: "Linux Title", Body: "Linux Body", Event: backend.EventStop})
	assert.NoError(t, err)
	assert.Equal(t, "notify-send", gotName)
	assert.Equal(t, "Linux Title", gotArgs[0])
	assert.Equal(t, "Linux Body", gotArgs[1])
}

func TestSendLinux_NotifySendMissing_ReturnsError(t *testing.T) {
	var called bool
	execCommand = func(name string, arg ...string) *exec.Cmd {
		called = true
		return exec.Command("true")
	}
	defer func() { execCommand = exec.Command }()

	lookPath = func(file string) (string, error) {
		return "", exec.ErrNotFound
	}
	defer func() { lookPath = exec.LookPath }()

	b := NewBackend()
	err := b.sendLinux(backend.Payload{Title: "Title", Body: "Body", Event: backend.EventStop})
	assert.Error(t, err, "expected error when notify-send missing")
	assert.ErrorContains(t, err, "system backend: notify-send not found")
	assert.False(t, called, "expected execCommand NOT to be called when notify-send is missing")
}

func TestSendDarwin_AllThreeEvents(t *testing.T) {
	cases := []struct {
		event backend.EventType
		title string
	}{
		{backend.EventStop, "Claude Finished"},
		{backend.EventNotification, "Claude Needs Input"},
		{backend.EventStopFailure, "Claude Error"},
	}

	for _, tc := range cases {
		t.Run(string(tc.event), func(t *testing.T) {
			var gotScript string
			execCommand = func(name string, arg ...string) *exec.Cmd {
				if len(arg) >= 2 {
					gotScript = arg[1]
				}
				return exec.Command("true")
			}
			defer func() { execCommand = exec.Command }()

			b := NewBackend()
			err := b.sendDarwin(backend.Payload{
				Title: tc.title,
				Body:  "some body",
				Event: tc.event,
			})
			assert.NoError(t, err, "sendDarwin returned error for event %s", tc.event)
			assert.True(t, strings.Contains(gotScript, tc.title), "expected title %q in osascript arg for event %s", tc.title, tc.event)
		})
	}
}
