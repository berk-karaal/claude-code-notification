package ntfy_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/berk-karaal/claude-code-notification/internal/backend"
	"github.com/berk-karaal/claude-code-notification/internal/backend/ntfy"
	"github.com/stretchr/testify/assert"
)

// Happy path testing
func Test_PostsCorrectHeadersAndBody(t *testing.T) {
	var gotMethod, gotTitle, gotBody, gotPath string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotTitle = r.Header.Get("Title")
		bodyBytes, _ := io.ReadAll(r.Body)
		gotBody = string(bodyBytes)
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	b, err := ntfy.NewBackend("test-topic", ts.URL)
	assert.NoError(t, err, "failed to create backend")

	payload := backend.Payload{
		Title:    "Claude Finished",
		Body:     "[myhost] Project: myproject",
		Event:    backend.EventStop,
		Hostname: "myhost",
		Project:  "myproject",
	}

	err = b.Send(payload)
	assert.NoError(t, err, "Send returned unexpected error")

	assert.Equal(t, "POST", gotMethod, "expected method POST")
	assert.Equal(t, "Claude Finished", gotTitle, "expected Title header 'Claude Finished'")
	assert.Equal(t, "[myhost] Project: myproject", gotBody, "expected body '[myhost] Project: myproject'")
	assert.Equal(t, "/test-topic", gotPath, "expected path '/test-topic'")
}

// Empty topic error
func Test_EmptyTopicReturnsError(t *testing.T) {
	_, err := ntfy.NewBackend("", "http://localhost")
	assert.Error(t, err, "expected error for empty topic, got nil")
}

// Empty server defaults to ntfy.sh
func Test_EmptyServerDefaultsToNtfySh(t *testing.T) {
	b, err := ntfy.NewBackend("test-topic", "")
	assert.NoError(t, err, "failed to create backend")
	assert.Equal(t, "https://ntfy.sh", b.Server, "expected default server 'https://ntfy.sh'")
}

// Test that server URL with trailing slashes is handled correctly (no double slash in path)
func TestSend_ServerTrailingSlashTrimmed(t *testing.T) {
	var gotPath string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	b, err := ntfy.NewBackend("mytopic", ts.URL+"////")
	assert.NoError(t, err, "failed to create backend")

	payload := backend.Payload{
		Title: "Claude Finished",
		Body:  "body",
		Event: backend.EventStop,
	}

	err = b.Send(payload)
	assert.NoError(t, err, "Send returned unexpected error")

	assert.Equal(t, "/mytopic", gotPath, "expected path '/mytopic' (no double slash)")
}

func TestSend_RequestCreationErrorIsReturned(t *testing.T) {
	b, err := ntfy.NewBackend("test-topic", "invalid :// url")
	assert.NoError(t, err, "failed to create backend")

	payload := backend.Payload{
		Title: "Claude Finished",
		Body:  "body",
		Event: backend.EventStop,
	}
	err = b.Send(payload)
	assert.Error(t, err, "expected error when creating request with invalid URL")
	assert.ErrorContains(t, err, "ntfy backend: failed to create request:")
}

func TestSend_RequestErrorIsReturned(t *testing.T) {
	b, err := ntfy.NewBackend("test-topic", "http://non-existent-server")
	assert.NoError(t, err, "failed to create backend")

	payload := backend.Payload{
		Title: "Claude Finished",
		Body:  "body",
		Event: backend.EventStop,
	}

	err = b.Send(payload)
	assert.Error(t, err, "expected error when sending to invalid server")
	assert.ErrorContains(t, err, "ntfy backend: request failed:")
}
