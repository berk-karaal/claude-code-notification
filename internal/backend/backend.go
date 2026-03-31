package backend

// EventType represents a Claude Code hook event type.
type EventType string

const (
	EventStop         EventType = "Stop"
	EventNotification EventType = "Notification"
	EventStopFailure  EventType = "StopFailure"
)

// Payload is the safe notification content derived from a hook event.
// It contains only non-sensitive fields: no prompts, paths, session IDs, or env vars.
type Payload struct {
	Title    string
	Body     string
	Event    EventType
	Hostname string
	Project  string
}

// Backend is implemented by each notification dispatcher.
// Send dispatches a notification. Implementations must return nil even on
// temporary failures (network errors, missing system tools) — they log to
// stderr and return nil so the binary always exits 0.
type Backend interface {
	Send(payload Payload) error
}
