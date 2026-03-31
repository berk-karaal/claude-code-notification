package ntfy

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/berk-karaal/claude-code-notification/internal/backend"
)

type Backend struct {
	Topic  string
	Server string
}

func NewBackend(topic string, server string) (*Backend, error) {
	if topic == "" {
		return nil, fmt.Errorf("ntfy backend: topic not configured")
	}

	if server == "" {
		server = "https://ntfy.sh"
	}

	return &Backend{
		Topic:  topic,
		Server: server,
	}, nil
}

// Send dispatches a notification to configured ntfy server via HTTP POST.
func (b *Backend) Send(payload backend.Payload) error {
	url := strings.TrimRight(b.Server, "/") + "/" + b.Topic

	req, err := http.NewRequest("POST", url, strings.NewReader(payload.Body))
	if err != nil {
		return fmt.Errorf("ntfy backend: failed to create request: %w", err)
	}
	req.Header.Set("Title", payload.Title)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ntfy backend: request failed: %w", err)
	}
	resp.Body.Close()

	return nil
}
