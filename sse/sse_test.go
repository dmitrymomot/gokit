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

func TestMessageWithTTL(t *testing.T) {
	t.Run("set ttl via constructor", func(t *testing.T) {
		ttl := 10 * time.Second
		msg := sse.NewMessage("test-event", "data", ttl)
		
		assert.Equal(t, ttl, msg.TTL)
		assert.Equal(t, "test-event", msg.Event)
		assert.Equal(t, "data", msg.Data)
		assert.False(t, msg.Timestamp.IsZero(), "Timestamp should be set")
	})

	t.Run("set ttl via WithTTL method", func(t *testing.T) {
		ttl := 5 * time.Minute
		msg := sse.NewMessage("test-event", "data").WithTTL(ttl)
		
		assert.Equal(t, ttl, msg.TTL)
		assert.Equal(t, "test-event", msg.Event)
		assert.Equal(t, "data", msg.Data)
	})
}

func TestMessageIsExpired(t *testing.T) {
	t.Run("message with no TTL never expires", func(t *testing.T) {
		msg := sse.NewMessage("test-event", "data")
		
		assert.False(t, msg.IsExpired(), "Message with no TTL should never expire")
	})

	t.Run("message with zero TTL never expires", func(t *testing.T) {
		msg := sse.NewMessage("test-event", "data").WithTTL(0)
		
		assert.False(t, msg.IsExpired(), "Message with zero TTL should never expire")
	})

	t.Run("message with negative TTL never expires", func(t *testing.T) {
		msg := sse.NewMessage("test-event", "data").WithTTL(-1 * time.Millisecond)
		
		assert.False(t, msg.IsExpired(), "Message with negative TTL should never expire")
	})

	t.Run("non-expired message", func(t *testing.T) {
		msg := sse.NewMessage("test-event", "data").WithTTL(100 * time.Millisecond)
		
		assert.False(t, msg.IsExpired(), "Message with future expiry should not be expired")
	})

	t.Run("expired message", func(t *testing.T) {
		// Create a message with short TTL
		msg := sse.NewMessage("test-event", "data").WithTTL(1 * time.Millisecond)
		
		// Wait for it to expire
		time.Sleep(5 * time.Millisecond)
		
		assert.True(t, msg.IsExpired(), "Message should be expired after TTL")
	})
}

func TestExpiredMessageSkipping(t *testing.T) {
	// Create a context with timeout to prevent test from hanging
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Create broker and server
	broker, err := memory.NewBroker()
	require.NoError(t, err)
	defer broker.Close()

	server, err := sse.NewServer(broker)
	require.NoError(t, err)
	defer server.Close()

	// Create test HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.ServeHTTP(w, r)
	}))
	defer ts.Close()
	
	// Setup client connection
	clientDone := make(chan struct{})
	clientMessages := make(chan string, 10)
	
	// Connect client and read messages
	go func() {
		defer close(clientDone)
		
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"?client_id=test-client", nil)
		if err != nil {
			t.Logf("Error creating request: %v", err)
			return
		}
		
		client := &http.Client{Timeout: time.Second}
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("Error connecting to server: %v", err)
			return
		}
		defer resp.Body.Close()
		
		buf := make([]byte, 1024)
		messageBuffer := ""
		
		readDone := make(chan struct{})
		go func() {
			defer close(readDone)
			for {
				select {
				case <-ctx.Done():
					return
				default:
					n, err := resp.Body.Read(buf)
					if err != nil {
						return
					}
					
					messageBuffer += string(buf[:n])
					messages := strings.Split(messageBuffer, "\n\n")
					
					for i, msg := range messages {
						if i < len(messages)-1 {
							if strings.TrimSpace(msg) != "" && !strings.HasPrefix(strings.TrimSpace(msg), ":") {
								select {
								case clientMessages <- strings.TrimSpace(msg):
								case <-ctx.Done():
									return
								}
							}
						}
					}
					
					if len(messages) > 0 {
						messageBuffer = messages[len(messages)-1]
					}
				}
			}
		}()
		
		select {
		case <-ctx.Done():
			return
		case <-readDone:
			return
		}
	}()
	
	// Wait for initial connection message
	var initialMessage string
	select {
	case initialMessage = <-clientMessages:
		require.Contains(t, initialMessage, "event: connected")
	case <-clientDone:
		t.Fatal("Client disconnected before receiving initial message")
	case <-ctx.Done():
		t.Fatal("Context deadline exceeded waiting for initial connection")
	}
	
	// Test 1: Send non-expired message
	nonExpiredMsg := sse.NewMessage("test-non-expired", map[string]string{
		"message": "This message should be received",
	})
	err = broker.Publish(ctx, nonExpiredMsg)
	require.NoError(t, err)
	
	// Verify non-expired message is received
	select {
	case msg := <-clientMessages:
		assert.Contains(t, msg, "event: test-non-expired")
		assert.Contains(t, msg, "This message should be received")
	case <-clientDone:
		t.Fatal("Client disconnected before receiving non-expired message")
	case <-ctx.Done():
		t.Fatal("Context deadline exceeded waiting for non-expired message")
	}
	
	// Test 2: Send already-expired message
	expiredMsg := sse.NewMessage("test-expired", map[string]string{
		"message": "This message should NOT be received",
	}).WithTTL(1 * time.Millisecond)
	
	// Wait for message to expire
	time.Sleep(5 * time.Millisecond)
	
	// Verify message is expired before sending
	assert.True(t, expiredMsg.IsExpired(), "Message should be expired before sending")
	
	// Send expired message
	err = broker.Publish(ctx, expiredMsg)
	require.NoError(t, err)
	
	// Brief wait to ensure message processing
	time.Sleep(50 * time.Millisecond)
	
	// Test 3: Send a message that expires shortly after sending
	expiringMsg := sse.NewMessage("test-expiring", map[string]string{
		"message": "This is another test",
	}).WithTTL(50 * time.Millisecond)
	
	err = broker.Publish(ctx, expiringMsg)
	require.NoError(t, err)
	
	// Wait to see if we get this message (should be delivered before it expires)
	select {
	case msg := <-clientMessages:
		assert.Contains(t, msg, "event: test-expiring")
	case <-clientDone:
		t.Fatal("Client disconnected before receiving message with short TTL")
	case <-ctx.Done():
		t.Fatal("Context deadline exceeded waiting for message with short TTL")
	}
	
	// Cancel the context to signal all goroutines to exit
	cancel()
	
	// Wait for client goroutine to finish
	select {
	case <-clientDone:
		// Client goroutine exited cleanly
	case <-time.After(100 * time.Millisecond):
		t.Log("Warning: Client goroutine did not exit in time")
	}
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
	case <-time.After(500 * time.Millisecond):
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
	case <-time.After(500 * time.Millisecond):
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
		case <-time.After(500 * time.Millisecond):
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
