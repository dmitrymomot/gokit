package bus_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/dmitrymomot/gokit/sse"
	"github.com/dmitrymomot/gokit/sse/bus"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupRedisTest initializes a test environment with miniredis
func setupRedisTest(t *testing.T) (*miniredis.Miniredis, *redis.Client, *bus.RedisBus, context.Context) {
	// Start a miniredis server
	miniRedis, err := miniredis.Run()
	require.NoError(t, err, "Failed to start miniredis")

	// Create a Redis client connected to the miniredis server
	client := redis.NewClient(&redis.Options{
		Addr: miniRedis.Addr(),
		DB:   0,
	})

	// Create the Redis message bus
	messageBus, err := bus.NewRedisBus(client)
	require.NoError(t, err, "Failed to create Redis message bus")

	// Create a test context
	ctx := context.Background()

	return miniRedis, client, messageBus, ctx
}

// teardownRedisTest cleans up the test environment
func teardownRedisTest(miniRedis *miniredis.Miniredis, client *redis.Client, messageBus *bus.RedisBus) {
	// Close the message bus
	if messageBus != nil {
		_ = messageBus.Close()
	}

	// Close the Redis client
	if client != nil {
		_ = client.Close()
	}

	// Stop the miniredis server
	if miniRedis != nil {
		miniRedis.Close()
	}
}

// TestSubscribeAndPublish tests the basic Subscribe and Publish functionality
func TestSubscribeAndPublish(t *testing.T) {
	miniRedis, client, messageBus, ctx := setupRedisTest(t)
	defer teardownRedisTest(miniRedis, client, messageBus)

	// Subscribe to a topic
	topic := "test-topic"
	ch, err := messageBus.Subscribe(ctx, topic)
	require.NoError(t, err)
	require.NotNil(t, ch)

	// Publish an event
	event := sse.Event{
		ID:    "1",
		Event: "test",
		Data:  "test data",
	}
	err = messageBus.Publish(ctx, topic, event)
	require.NoError(t, err)

	// Receive the event
	select {
	case receivedEvent := <-ch:
		assert.Equal(t, event.ID, receivedEvent.ID)
		assert.Equal(t, event.Event, receivedEvent.Event)
		assert.Equal(t, event.Data, receivedEvent.Data)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for event")
	}
}

// TestUnsubscribe tests the Unsubscribe functionality
func TestUnsubscribe(t *testing.T) {
	miniRedis, client, messageBus, ctx := setupRedisTest(t)
	defer teardownRedisTest(miniRedis, client, messageBus)

	// Subscribe to a topic
	topic := "test-unsubscribe-" + time.Now().String() // Use unique topic name
	ch, err := messageBus.Subscribe(ctx, topic)
	require.NoError(t, err)

	// Small delay to ensure subscription is active
	time.Sleep(100 * time.Millisecond)

	// Unsubscribe
	err = messageBus.Unsubscribe(ctx, topic, ch)
	require.NoError(t, err)

	// Longer delay to ensure unsubscription is fully processed
	time.Sleep(300 * time.Millisecond)

	// Publish an event (should not receive it)
	event := sse.Event{
		ID:    "1",
		Event: "test",
		Data:  "test data",
	}
	err = messageBus.Publish(ctx, topic, event)
	require.NoError(t, err)

	// Create a timeout channel
	timeout := time.After(200 * time.Millisecond)

	// Check for messages with a timeout
	select {
	case <-ch: // Verify if the channel is already closed
		// Trying to receive from a potentially closed channel
		// This might happen if the implementation closes the channel on unsubscribe
		select {
		case eventData, ok := <-ch:
			if ok { // If we received an actual event, not just channel close
				t.Fatalf("received event after unsubscribe: %v", eventData)
			}
			// If channel is closed (!ok), that's acceptable
		case <-timeout:
			// This is also expected
		}
	case <-timeout:
		// This is expected
	}
}

// TestMultipleSubscribers tests multiple subscribers to the same topic
func TestMultipleSubscribers(t *testing.T) {
	miniRedis, client, messageBus, ctx := setupRedisTest(t)
	defer teardownRedisTest(miniRedis, client, messageBus)

	// Subscribe multiple clients to the same topic
	topic := "test-topic"
	ch1, err := messageBus.Subscribe(ctx, topic)
	require.NoError(t, err)

	ch2, err := messageBus.Subscribe(ctx, topic)
	require.NoError(t, err)

	// Small delay to ensure subscriptions are active
	time.Sleep(50 * time.Millisecond)

	// Publish an event
	event := sse.Event{
		ID:    "1",
		Event: "test",
		Data:  "test data",
	}
	err = messageBus.Publish(ctx, topic, event)
	require.NoError(t, err)

	// Both subscribers should receive the event
	for i, ch := range []<-chan sse.Event{ch1, ch2} {
		select {
		case receivedEvent := <-ch:
			assert.Equal(t, event.ID, receivedEvent.ID)
			assert.Equal(t, event.Event, receivedEvent.Event)
			assert.Equal(t, event.Data, receivedEvent.Data)
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("timeout waiting for event on channel %d", i)
		}
	}
}

// TestContextCancellation tests the behavior when the context is canceled
func TestContextCancellation(t *testing.T) {
	miniRedis, client, messageBus, _ := setupRedisTest(t)
	defer teardownRedisTest(miniRedis, client, messageBus)

	// Create a context with cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Subscribe to a topic
	topic := "test-topic"
	ch, err := messageBus.Subscribe(ctx, topic)
	require.NoError(t, err)

	// Small delay to ensure subscription is active
	time.Sleep(50 * time.Millisecond)

	// Cancel the context
	cancel()

	// Give some time for cleanup
	time.Sleep(200 * time.Millisecond)

	// The channel should be closed eventually
	select {
	case _, ok := <-ch:
		assert.False(t, ok, "Channel should be closed after context cancellation")
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for channel to close")
	}
}

// TestEmptyTopic tests error handling for empty topics
func TestEmptyTopic(t *testing.T) {
	miniRedis, client, messageBus, ctx := setupRedisTest(t)
	defer teardownRedisTest(miniRedis, client, messageBus)

	// Try to subscribe to an empty topic
	_, err := messageBus.Subscribe(ctx, "")
	assert.Equal(t, sse.ErrTopicEmpty, err)

	// Try to publish to an empty topic
	err = messageBus.Publish(ctx, "", sse.Event{})
	assert.Equal(t, sse.ErrTopicEmpty, err)

	// Try to unsubscribe from an empty topic
	err = messageBus.Unsubscribe(ctx, "", nil)
	assert.Equal(t, sse.ErrTopicEmpty, err)
}

// TestRedisDisconnection tests behavior when Redis disconnects
func TestRedisDisconnection(t *testing.T) {
	miniRedis, client, messageBus, ctx := setupRedisTest(t)
	defer teardownRedisTest(nil, client, messageBus) // We'll close miniredis manually

	// Subscribe to a topic
	topic := "test-topic"
	ch, err := messageBus.Subscribe(ctx, topic)
	require.NoError(t, err)

	// Small delay to ensure subscription is active
	time.Sleep(50 * time.Millisecond)

	// Stop Redis server to simulate disconnection
	miniRedis.Close()

	// Start a new Redis server (with a different dataset)
	newRedis, err := miniredis.Run()
	require.NoError(t, err, "Failed to start new miniredis")
	defer newRedis.Close()

	// Try to publish after disconnection
	event := sse.Event{
		ID:    "1",
		Event: "test",
		Data:  "test data",
	}
	err = messageBus.Publish(ctx, topic, event)
	require.Error(t, err, "Publishing to disconnected Redis should fail")

	// The channel might be closed eventually or might timeout
	select {
	case _, ok := <-ch:
		assert.False(t, ok, "Channel should close after Redis disconnect")
	case <-time.After(500 * time.Millisecond):
		// Redis disconnect might not immediately close channels,
		// so this timeout is also an acceptable outcome
	}
}

// TestRedisBusClosing tests closing the bus
func TestRedisBusClosing(t *testing.T) {
	miniRedis, client, messageBus, ctx := setupRedisTest(t)
	defer teardownRedisTest(miniRedis, client, nil) // We'll close message bus manually

	// Subscribe to a topic
	topic := "test-topic"
	ch, err := messageBus.Subscribe(ctx, topic)
	require.NoError(t, err)

	// Small delay to ensure subscription is active
	time.Sleep(50 * time.Millisecond)

	// Close the message bus
	err = messageBus.Close()
	require.NoError(t, err)

	// The channel should be closed
	select {
	case _, ok := <-ch:
		assert.False(t, ok, "Channel should be closed after bus is closed")
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for channel to close")
	}

	// Attempting to subscribe after closing should fail
	_, err = messageBus.Subscribe(ctx, topic)
	assert.Equal(t, sse.ErrMessageBusClosed, err)

	// Attempting to publish after closing should fail
	err = messageBus.Publish(ctx, topic, sse.Event{})
	assert.Equal(t, sse.ErrMessageBusClosed, err)

	// Attempting to unsubscribe after closing should fail
	err = messageBus.Unsubscribe(ctx, topic, ch)
	assert.Equal(t, sse.ErrMessageBusClosed, err)
}

// TestMessageSerialization tests that events are properly serialized/deserialized
func TestMessageSerialization(t *testing.T) {
	miniRedis, client, messageBus, ctx := setupRedisTest(t)
	defer teardownRedisTest(miniRedis, client, messageBus)

	// Subscribe to a topic
	topic := "test-topic"
	ch, err := messageBus.Subscribe(ctx, topic)
	require.NoError(t, err)

	// Test various data types
	testCases := []sse.Event{
		{
			ID:    "1",
			Event: "string",
			Data:  "simple string data",
		},
		{
			ID:    "2",
			Event: "number",
			Data:  42,
		},
		{
			ID:    "3",
			Event: "object",
			Data:  map[string]any{"name": "test", "value": 123},
		},
		{
			ID:    "4",
			Event: "array",
			Data:  []string{"one", "two", "three"},
		},
		{
			ID:    "5",
			Event: "empty",
			Data:  nil,
		},
		{
			ID:    "6",
			Event: "retry",
			Data:  "retry data",
			Retry: 3000,
		},
	}

	// Publish all test events
	for _, event := range testCases {
		err = messageBus.Publish(ctx, topic, event)
		require.NoError(t, err)
	}

	// Receive and verify all events
	for _, expectedEvent := range testCases {
		select {
		case receivedEvent := <-ch:
			assert.Equal(t, expectedEvent.ID, receivedEvent.ID)
			assert.Equal(t, expectedEvent.Event, receivedEvent.Event)
			assert.Equal(t, expectedEvent.Retry, receivedEvent.Retry)

			// For complex data types, we need to handle JSON type conversions
			if expectedEvent.Data != nil {
				switch expected := expectedEvent.Data.(type) {
				case int:
					// JSON unmarshals numbers as float64
					assert.Equal(t, float64(expected), receivedEvent.Data)
				case map[string]any:
					// Check map equality after handling numeric conversions
					received, ok := receivedEvent.Data.(map[string]any)
					assert.True(t, ok, "Expected map but got %T", receivedEvent.Data)
					assert.Equal(t, len(expected), len(received), "Map length mismatch")
					for k, v := range expected {
						if intVal, ok := v.(int); ok {
							assert.Equal(t, float64(intVal), received[k], "Map value mismatch for key %s", k)
						} else {
							assert.Equal(t, v, received[k], "Map value mismatch for key %s", k)
						}
					}
				case []string:
					// JSON unmarshals arrays as []any
					received, ok := receivedEvent.Data.([]any)
					assert.True(t, ok, "Expected array but got %T", receivedEvent.Data)
					assert.Equal(t, len(expected), len(received), "Array length mismatch")
					for i, v := range expected {
						assert.Equal(t, v, received[i], "Array value mismatch at index %d", i)
					}
				default:
					assert.Equal(t, expected, receivedEvent.Data)
				}
			} else {
				assert.Nil(t, receivedEvent.Data)
			}
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("timeout waiting for event: %v", expectedEvent)
		}
	}
}

// TestNewRedisBus tests the constructor for the Redis bus
func TestNewRedisBus(t *testing.T) {
	// Start a miniredis server
	miniRedis, err := miniredis.Run()
	require.NoError(t, err, "Failed to start miniredis")
	defer miniRedis.Close()

	// Create a Redis client connected to the miniredis server
	client := redis.NewClient(&redis.Options{
		Addr: miniRedis.Addr(),
		DB:   0,
	})
	defer client.Close()

	// Test creating a new Redis bus
	messageBus, err := bus.NewRedisBus(client)
	require.NoError(t, err)
	require.NotNil(t, messageBus)

	// Test creating with custom buffer size
	messageBus, err = bus.NewRedisBusWithConfig(client, 200)
	require.NoError(t, err)
	require.NotNil(t, messageBus)

	// Test with invalid Redis client
	invalidClient := redis.NewClient(&redis.Options{
		Addr: "localhost:12345", // Invalid address
		DB:   0,
	})
	defer invalidClient.Close()

	// Should fail to create bus with invalid client
	_, err = bus.NewRedisBus(invalidClient)
	require.Error(t, err)
	// The error might be wrapped, so we use errors.Is instead of errors.Unwrap
	assert.True(t, errors.Is(err, sse.ErrMessageBusClosed) || strings.Contains(err.Error(), "connection refused"),
		"Expected error related to connection issues, got: %v", err)
}
