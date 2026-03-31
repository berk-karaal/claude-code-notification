package integration

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// TestNtfy_StopEvent_PostsToServer verifies the binary sends an HTTP POST to the
// configured ntfy server with correct title and body.
func TestNtfy_StopEvent_PostsToServer(t *testing.T) {
	var mu sync.Mutex
	var gotMethod, gotTitle, gotBody string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		gotMethod = r.Method
		gotTitle = r.Header.Get("Title")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cwd := setupProjectDir(t, `{
		"system_notification": {"enabled": false},
		"ntfy": {"enabled": true, "topic": "test-topic", "server": "`+srv.URL+`"}
	}`)

	_, stderr, code := runBinary(t, makeInput("Stop", cwd))
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, stderr)
	}

	mu.Lock()
	defer mu.Unlock()
	if gotMethod != "POST" {
		t.Errorf("expected POST request, got %q", gotMethod)
	}
	if gotTitle != "Claude Finished" {
		t.Errorf("expected title 'Claude Finished', got %q", gotTitle)
	}
	if !strings.Contains(gotBody, "Project:") {
		t.Errorf("expected body to contain 'Project:', got %q", gotBody)
	}
}

// TestNtfy_NotificationEvent_PostsCorrectTitle verifies the Notification event
// produces a "Claude Needs Input" title.
func TestNtfy_NotificationEvent_PostsCorrectTitle(t *testing.T) {
	var mu sync.Mutex
	var gotTitle string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		gotTitle = r.Header.Get("Title")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cwd := setupProjectDir(t, `{
		"system_notification": {"enabled": false},
		"ntfy": {"enabled": true, "topic": "test-topic", "server": "`+srv.URL+`"}
	}`)

	_, stderr, code := runBinary(t, makeInput("Notification", cwd))
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, stderr)
	}

	mu.Lock()
	defer mu.Unlock()
	if gotTitle != "Claude Needs Input" {
		t.Errorf("expected title 'Claude Needs Input', got %q", gotTitle)
	}
}

// TestNtfy_NoTopic_SkipsSend verifies that when ntfy is enabled but topic is
// empty, the binary logs a skip and exits 0 without making an HTTP request.
func TestNtfy_NoTopic_SkipsSend(t *testing.T) {
	var mu sync.Mutex
	var called bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cwd := setupProjectDir(t, `{
		"system_notification": {"enabled": false},
		"ntfy": {"enabled": true, "topic": "", "server": "`+srv.URL+`"}
	}`)

	_, stderr, code := runBinary(t, makeInput("Stop", cwd))
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, stderr)
	}
	if !strings.Contains(stderr, "topic not configured") {
		t.Errorf("expected stderr to contain 'topic not configured', got %q", stderr)
	}

	mu.Lock()
	defer mu.Unlock()
	if called {
		t.Error("expected no HTTP request when topic is empty")
	}
}
