package sse

import "errors"

var (
	// ErrClientClosed is returned when trying to send to a closed client
	ErrClientClosed = errors.New("client is closed")

	// ErrServerClosed is returned when trying to use a closed server
	ErrServerClosed = errors.New("server is closed")

	// ErrTopicEmpty is returned when the topic is empty
	ErrTopicEmpty = errors.New("topic cannot be empty")

	// ErrMessageEmpty is returned when the message is empty
	ErrMessageEmpty = errors.New("message cannot be empty")

	// ErrInvalidEventID is returned when the event ID is invalid
	ErrInvalidEventID = errors.New("invalid event ID")

	// ErrMessageBusClosed is returned when the message bus is closed
	ErrMessageBusClosed = errors.New("message bus is closed")

	// ErrNoFlusher is returned when the ResponseWriter does not implement http.Flusher
	ErrNoFlusher = errors.New("response writer does not implement http.Flusher")
)
