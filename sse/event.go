package sse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Event represents a Server-Sent Event
type Event struct {
	// ID is the event ID
	ID string

	// Event is the event type
	Event string

	// Data is the event data (can be any type)
	Data any

	// Retry is the reconnection time in milliseconds
	Retry int
}

// String converts an Event to its SSE string representation
func (e Event) String() string {
	var buf bytes.Buffer

	// Add ID field if present
	if e.ID != "" {
		fmt.Fprintf(&buf, "id: %s\n", e.ID)
	}

	// Add Event field if present
	if e.Event != "" {
		fmt.Fprintf(&buf, "event: %s\n", e.Event)
	}

	// Add Data field, handling different data types
	if e.Data != nil {
		strData, err := convertDataToString(e.Data)
		if err == nil && strData != "" {
			for line := range strings.SplitSeq(strData, "\n") {
				fmt.Fprintf(&buf, "data: %s\n", line)
			}
		}
	}

	// Add Retry field if present
	if e.Retry > 0 {
		fmt.Fprintf(&buf, "retry: %d\n", e.Retry)
	}

	// End with a blank line
	fmt.Fprint(&buf, "\n")

	return buf.String()
}

// Write writes the event to the given writer in SSE format
func (e Event) Write(w io.Writer) error {
	_, err := fmt.Fprint(w, e.String())
	return err
}

// convertDataToString converts the event data to a string representation
func convertDataToString(data any) (string, error) {
	switch v := data.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case error:
		return v.Error(), nil
	case fmt.Stringer:
		return v.String(), nil
	default:
		// For all other types, use JSON marshaling
		jsonData, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(jsonData), nil
	}
}
