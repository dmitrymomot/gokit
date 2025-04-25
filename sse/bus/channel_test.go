package bus_test

import (
	"context"
	"testing"
	"time"

	"github.com/dmitrymomot/gokit/sse"
	"github.com/dmitrymomot/gokit/sse/bus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChannelBus(t *testing.T) {
	t.Run("Subscribe and Publish", func(t *testing.T) {
		// Create a new channel bus
		bus := bus.NewChannelBus()
		ctx := context.Background()

		// Subscribe to a topic
		topic := "test-topic"
		ch, err := bus.Subscribe(ctx, topic)
		require.NoError(t, err)
		require.NotNil(t, ch)

		// Publish an event
		event := sse.Event{
			ID:    "1",
			Event: "test",
			Data:  "test data",
		}
		err = bus.Publish(ctx, topic, event)
		require.NoError(t, err)

		// Receive the event
		select {
		case receivedEvent := <-ch:
			assert.Equal(t, event.ID, receivedEvent.ID)
			assert.Equal(t, event.Event, receivedEvent.Event)
			assert.Equal(t, event.Data, receivedEvent.Data)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("timeout waiting for event")
		}

		// Close the bus
		err = bus.Close()
		require.NoError(t, err)
	})

	t.Run("Unsubscribe", func(t *testing.T) {
		// Create a new channel bus
		bus := bus.NewChannelBus()
		ctx := context.Background()

		// Subscribe to a topic
		topic := "test-topic"
		ch, err := bus.Subscribe(ctx, topic)
		require.NoError(t, err)

		// Unsubscribe
		err = bus.Unsubscribe(ctx, topic, ch)
		require.NoError(t, err)

		// Publish an event (should not receive it)
		event := sse.Event{
			ID:    "1",
			Event: "test",
			Data:  "test data",
		}
		err = bus.Publish(ctx, topic, event)
		require.NoError(t, err)

		// Should not receive any events
		select {
		case <-ch:
			t.Fatal("received event after unsubscribe")
		case <-time.After(100 * time.Millisecond):
			// This is expected
		}

		// Close the bus
		err = bus.Close()
		require.NoError(t, err)
	})

	t.Run("Multiple Subscribers", func(t *testing.T) {
		// Create a new channel bus
		bus := bus.NewChannelBus()
		ctx := context.Background()

		// Subscribe multiple clients to the same topic
		topic := "test-topic"
		ch1, err := bus.Subscribe(ctx, topic)
		require.NoError(t, err)

		ch2, err := bus.Subscribe(ctx, topic)
		require.NoError(t, err)

		// Publish an event
		event := sse.Event{
			ID:    "1",
			Event: "test",
			Data:  "test data",
		}
		err = bus.Publish(ctx, topic, event)
		require.NoError(t, err)

		// Both subscribers should receive the event
		for _, ch := range []<-chan sse.Event{ch1, ch2} {
			select {
			case receivedEvent := <-ch:
				assert.Equal(t, event.ID, receivedEvent.ID)
				assert.Equal(t, event.Event, receivedEvent.Event)
				assert.Equal(t, event.Data, receivedEvent.Data)
			case <-time.After(100 * time.Millisecond):
				t.Fatal("timeout waiting for event")
			}
		}

		// Close the bus
		err = bus.Close()
		require.NoError(t, err)
	})

	t.Run("Context Cancellation", func(t *testing.T) {
		// Create a new channel bus
		bus := bus.NewChannelBus()

		// Create a context with cancel
		ctx, cancel := context.WithCancel(context.Background())

		// Subscribe to a topic
		topic := "test-topic"
		_, err := bus.Subscribe(ctx, topic)
		require.NoError(t, err)

		// Cancel the context
		cancel()

		// Give some time for cleanup
		time.Sleep(10 * time.Millisecond)

		// The subscription should be removed
		// We can verify this by checking if a new publish would work without errors
		err = bus.Publish(context.Background(), topic, sse.Event{ID: "2"})
		require.NoError(t, err)

		// Close the bus
		err = bus.Close()
		require.NoError(t, err)
	})

	t.Run("Empty Topic", func(t *testing.T) {
		// Create a new channel bus
		bus := bus.NewChannelBus()
		ctx := context.Background()

		// Try to subscribe to an empty topic
		_, err := bus.Subscribe(ctx, "")
		assert.Equal(t, sse.ErrTopicEmpty, err)

		// Try to publish to an empty topic
		err = bus.Publish(ctx, "", sse.Event{})
		assert.Equal(t, sse.ErrTopicEmpty, err)

		// Try to unsubscribe from an empty topic
		err = bus.Unsubscribe(ctx, "", nil)
		assert.Equal(t, sse.ErrTopicEmpty, err)

		// Close the bus
		err = bus.Close()
		require.NoError(t, err)
	})
}
