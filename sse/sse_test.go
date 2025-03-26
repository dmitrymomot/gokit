package sse_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dmitrymomot/gokit/sse"
	"github.com/dmitrymomot/gokit/sse/brokers/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMessage(t *testing.T) {
	// Create a new message
	msg := sse.NewMessage("test-event", map[string]string{"hello": "world"})

	// Check values
	assert.Equal(t, "test-event", msg.Event)
	assert.Equal(t, map[string]string{"hello": "world"}, msg.Data)
	assert.False(t, msg.Timestamp.IsZero(), "Timestamp should be set")
}

func TestMessageChaining(t *testing.T) {
	// Test builder pattern/method chaining
	msg := sse.NewMessage("test-event", "data").
		ForClient("client-123").
		ForChannel("channel-abc").
		WithID("msg-456")

	assert.Equal(t, "test-event", msg.Event)
	assert.Equal(t, "data", msg.Data)
	assert.Equal(t, "client-123", msg.ClientID)
	assert.Equal(t, "channel-abc", msg.Channel)
	assert.Equal(t, "msg-456", msg.ID)
}

func TestMessageValidate(t *testing.T) {
	t.Run("valid message", func(t *testing.T) {
		msg := sse.NewMessage("test-event", "data")
		err := msg.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing event", func(t *testing.T) {
		msg := sse.Message{Data: "data"}
		err := msg.Validate()
		assert.Error(t, err)
		assert.ErrorIs(t, err, sse.ErrInvalidMessage)
	})

	t.Run("missing data", func(t *testing.T) {
		msg := sse.Message{Event: "event"}
		err := msg.Validate()
		assert.Error(t, err)
		assert.ErrorIs(t, err, sse.ErrInvalidMessage)
	})
}

func TestMessageToEventString(t *testing.T) {
	// Create test message
	msg := sse.NewMessage("test-event", map[string]string{"key": "value"}).WithID("123")

	// Convert to event string
	eventStr, err := msg.ToEventString()
	require.NoError(t, err)

	// Check format
	assert.Contains(t, eventStr, "id: 123\n")
	assert.Contains(t, eventStr, "event: test-event\n")
	assert.Contains(t, eventStr, "data: ")
	
	// Verify data is JSON-encoded
	assert.Contains(t, eventStr, `{"key":"value"}`)
	
	// Should end with double newline
	assert.True(t, strings.HasSuffix(eventStr, "\n\n"), "Should end with double newline")
}

func TestNewServer(t *testing.T) {
	t.Run("create server with valid broker", func(t *testing.T) {
		broker, err := memory.NewBroker()
		require.NoError(t, err)
		defer broker.Close()

		server, err := sse.NewServer(broker)
		require.NoError(t, err)
		require.NotNil(t, server)
		
		// Clean up
		err = server.Close()
		assert.NoError(t, err)
	})

	t.Run("error with nil broker", func(t *testing.T) {
		server, err := sse.NewServer(nil)
		assert.Error(t, err)
		assert.Nil(t, server)
		assert.ErrorIs(t, err, sse.ErrNoBrokerProvided)
	})

	t.Run("with custom keep-alive interval", func(t *testing.T) {
		broker, err := memory.NewBroker()
		require.NoError(t, err)
		defer broker.Close()

		interval := 5 * time.Second
		server, err := sse.NewServer(broker, sse.WithKeepAliveInterval(interval))
		require.NoError(t, err)
		require.NotNil(t, server)
		
		// Clean up
		err = server.Close()
		assert.NoError(t, err)
	})
}

func TestServerServeHTTP(t *testing.T) {
	t.Run("method not allowed", func(t *testing.T) {
		broker, err := memory.NewBroker()
		require.NoError(t, err)
		defer broker.Close()

		server, err := sse.NewServer(broker)
		require.NoError(t, err)
		defer server.Close()

		// Create test request with POST method
		req := httptest.NewRequest(http.MethodPost, "/events", nil)
		w := httptest.NewRecorder()

		// Serve the request
		server.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("missing client_id", func(t *testing.T) {
		broker, err := memory.NewBroker()
		require.NoError(t, err)
		defer broker.Close()

		server, err := sse.NewServer(broker)
		require.NoError(t, err)
		defer server.Close()

		// Create test request without client_id
		req := httptest.NewRequest(http.MethodGet, "/events", nil)
		w := httptest.NewRecorder()

		// Serve the request
		server.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Missing client_id")
	})
}

func TestClientServerCommunication(t *testing.T) {
	// Create broker and server
	broker, err := memory.NewBroker()
	require.NoError(t, err)
	defer broker.Close()

	server, err := sse.NewServer(broker, sse.WithKeepAliveInterval(1*time.Second))
	require.NoError(t, err)
	defer server.Close()

	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.ServeHTTP(w, r)
	}))
	defer ts.Close()

	// Setup client connection in a goroutine
	clientDone := make(chan struct{})
	clientMessages := make(chan string, 10)

	// This simulates a client connection with parsing of SSE messages
	go func() {
		clientID := "test-client"
		req, err := http.NewRequest(http.MethodGet, ts.URL+"?client_id="+clientID, nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Read from the SSE stream
		buf := make([]byte, 1024)
		messageBuffer := ""

		for {
			select {
			case <-clientDone:
				return
			default:
				n, err := resp.Body.Read(buf)
				if err != nil {
					// Connection closed or error
					close(clientDone)
					return
				}

				messageBuffer += string(buf[:n])
				
				// Simple SSE parsing - split by double newline
				messages := strings.Split(messageBuffer, "\n\n")
				for i, msg := range messages {
					if i < len(messages)-1 {
						// Complete message
						if strings.TrimSpace(msg) != "" && !strings.HasPrefix(strings.TrimSpace(msg), ":") {
							clientMessages <- strings.TrimSpace(msg)
						}
					}
				}

				// Keep the incomplete part
				if len(messages) > 0 {
					messageBuffer = messages[len(messages)-1]
				}
			}
		}
	}()

	// Wait for client to connect and receive the initial connection event
	var initialMessage string
	select {
	case initialMessage = <-clientMessages:
		// Received the initial connection message
		assert.Contains(t, initialMessage, "event: connected")
		assert.Contains(t, initialMessage, "client_id")
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for initial connection event")
	}

	// Send a test message through the broker
	ctx := context.Background()
	testMessage := sse.NewMessage("test-event", map[string]string{"foo": "bar"})
	err = broker.Publish(ctx, testMessage)
	require.NoError(t, err)

	// Wait for message to be received
	var receivedMessage string
	select {
	case receivedMessage = <-clientMessages:
		// Received a message
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for message")
	}

	// Close client connection
	close(clientDone)

	// Verify received message
	assert.Contains(t, receivedMessage, "event: test-event")
	assert.Contains(t, receivedMessage, `data: {"foo":"bar"}`)
}

func TestMemoryBroker(t *testing.T) {
	t.Run("publish and subscribe", func(t *testing.T) {
		// Create broker
		broker, err := memory.NewBroker()
		require.NoError(t, err)
		defer broker.Close()

		// Subscribe to messages
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		msgCh, err := broker.Subscribe(ctx)
		require.NoError(t, err)

		// Publish a message
		testMsg := sse.NewMessage("test-event", "test-data")
		err = broker.Publish(ctx, testMsg)
		require.NoError(t, err)

		// Wait for message
		var receivedMsg sse.Message
		select {
		case receivedMsg = <-msgCh:
			// Message received
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for message")
		}

		// Verify message
		assert.Equal(t, "test-event", receivedMsg.Event)
		assert.Equal(t, "test-data", receivedMsg.Data)
	})

	t.Run("validate message on publish", func(t *testing.T) {
		broker, err := memory.NewBroker()
		require.NoError(t, err)
		defer broker.Close()

		ctx := context.Background()
		
		// Invalid message - no event
		invalidMsg := sse.Message{Data: "data"}
		err = broker.Publish(ctx, invalidMsg)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sse.ErrInvalidMessage)
	})

	t.Run("close broker", func(t *testing.T) {
		broker, err := memory.NewBroker()
		require.NoError(t, err)

		ctx := context.Background()
		
		// Subscribe first
		msgCh, err := broker.Subscribe(ctx)
		require.NoError(t, err)

		// Close broker
		err = broker.Close()
		require.NoError(t, err)

		// Publishing should fail
		msg := sse.NewMessage("test", "data")
		err = broker.Publish(ctx, msg)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sse.ErrBrokerClosed)

		// Channel should be closed
		_, ok := <-msgCh
		assert.False(t, ok, "Channel should be closed")
	})
}
