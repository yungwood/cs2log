package cs2log

import "time"

// Event is a parsed CS2 server log event.
type Event interface {
	EventType() string
	Timestamp() time.Time
	// RawLine returns retained source text for the event. It may contain player
	// names or operational data; event types that can carry secrets should
	// redact them before exposing raw text.
	RawLine() string
}

// BaseEvent contains fields common to every parsed event.
type BaseEvent struct {
	Type    string    `json:"type"`
	TimeUTC time.Time `json:"time"`
	// Raw is the retained source line or multiline block text for this event.
	// It may be redacted for event types that can contain sensitive values.
	Raw string `json:"raw"`
}

// EventType returns the event type name.
func (e BaseEvent) EventType() string { return e.Type }

// Timestamp returns the parsed event timestamp in UTC.
func (e BaseEvent) Timestamp() time.Time { return e.TimeUTC }

// RawLine returns the retained raw log line or multiline block text.
func (e BaseEvent) RawLine() string { return e.Raw }

// Unknown is emitted when a line has a valid CS2 timestamp but no known event
// pattern matches its payload.
type Unknown struct {
	BaseEvent
	Payload string `json:"payload"`
}

// Ignored is emitted for known log lines that are intentionally treated as
// non-domain noise.
type Ignored struct {
	BaseEvent
	Reason  string `json:"reason"`
	Payload string `json:"payload"`
}
