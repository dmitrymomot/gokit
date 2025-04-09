package sse

import (
	"bytes"
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

	// Data is the event data
	Data string

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

	// Add Data field, handling multiline data
	if e.Data != "" {
		for line := range strings.SplitSeq(e.Data, "\n") {
			fmt.Fprintf(&buf, "data: %s\n", line)
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
