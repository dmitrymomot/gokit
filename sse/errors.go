package sse

import "errors"

var (
	// ErrNoBrokerProvided is returned when attempting to create a server without a broker
	ErrNoBrokerProvided = errors.New("no broker provided")

	// ErrClientNotConnected is returned when attempting to send to a non-existent client
	ErrClientNotConnected = errors.New("client not connected")

	// ErrServerClosed is returned when attempting to use a closed server
	ErrServerClosed = errors.New("server closed")

	// ErrInvalidMessage is returned when a message is invalid
	ErrInvalidMessage = errors.New("invalid message")

	// ErrBrokerClosed is returned when the broker is closed
	ErrBrokerClosed = errors.New("broker closed")

	// ErrClientAlreadyExists is returned when a client with the same ID already exists
	ErrClientAlreadyExists = errors.New("client already exists")
)
