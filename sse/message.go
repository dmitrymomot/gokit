package sse

import (
	"encoding/json"
	"fmt"
	"time"
)

// Message represents a Server-Sent Event message
type Message struct {
	// ID is an optional identifier for the message
	ID string `json:"id,omitempty"`

	// Event is the event type/name
	Event string `json:"event,omitempty"`

	// Data is the message payload
	Data any `json:"data,omitempty"`

	// ClientID indicates a specific client this message targets
	// If empty, the message is considered a broadcast
	ClientID string `json:"client_id,omitempty"`

	// Channel indicates a specific channel this message targets
	// If empty, and ClientID is also empty, the message is broadcast to all clients
	Channel string `json:"channel,omitempty"`

	// Timestamp is when the message was created
	Timestamp time.Time `json:"timestamp"`

	// TTL is the time-to-live duration for this message
	// If zero, the message never expires
	// If set, the message will be skipped if (now > Timestamp + TTL)
	TTL time.Duration `json:"ttl,omitempty"`
}

// NewMessage creates a new message with the current timestamp
func NewMessage(event string, data any, ttl ...time.Duration) Message {
	m := Message{
		Event:     event,
		Data:      data,
		Timestamp: time.Now(),
	}
	if len(ttl) > 0 {
		m.TTL = ttl[0]
	}
	return m
}

// ForClient sets the target client ID for the message
func (m Message) ForClient(clientID string) Message {
	m.ClientID = clientID
	return m
}

// ForChannel sets the target channel for the message
func (m Message) ForChannel(channel string) Message {
	m.Channel = channel
	return m
}

// WithID sets the message ID
func (m Message) WithID(id string) Message {
	m.ID = id
	return m
}

// WithTTL sets a time-to-live duration for the message
// After this duration from the timestamp, the message is considered expired
// and will not be delivered to clients
func (m Message) WithTTL(duration time.Duration) Message {
	m.TTL = duration
	return m
}

// IsExpired checks if the message has exceeded its TTL
// If TTL is zero, the message never expires and this returns false
func (m Message) IsExpired() bool {
	if m.TTL <= 0 {
		return false
	}
	return time.Since(m.Timestamp) > m.TTL
}

// Validate checks if the message is valid
func (m Message) Validate() error {
	if m.Event == "" {
		return fmt.Errorf("%w: event name cannot be empty", ErrInvalidMessage)
	}

	if m.Data == nil {
		return fmt.Errorf("%w: data cannot be nil", ErrInvalidMessage)
	}

	return nil
}

// ToEventString formats the message as an SSE event string
func (m Message) ToEventString() (string, error) {
	var result string

	// Add ID field if present
	if m.ID != "" {
		result += fmt.Sprintf("id: %s\n", m.ID)
	}

	// Add event type
	result += fmt.Sprintf("event: %s\n", m.Event)

	// Add data (must be marshaled to JSON)
	data, err := json.Marshal(m.Data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}
	result += fmt.Sprintf("data: %s\n", data)

	// End with an extra newline to complete the event
	result += "\n"

	return result, nil
}
