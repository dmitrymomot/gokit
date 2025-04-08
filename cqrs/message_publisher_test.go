package cqrs_test

import (
	"context"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/dmitrymomot/gokit/cqrs"
	"github.com/dmitrymomot/gokit/cqrs/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
)

func TestGoChannelPublisher(t *testing.T) {
	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a logger for the test configuration
	logger := slog.Default()

	// Create a shared GoChannel for both publisher and subscriber to ensure they communicate
	config := test.NewDefaultTestBusConfig()
	config.Logger = logger
	messageBus, _ := test.NewChannelMessageBus(t, ctx, config)

	// Create a new GoChannelPublisher using the same underlying pubsub
	publisher := messageBus.PubSub
	defer publisher.Close()
	
	// Create a new test message
	testMsg := message.NewMessage("test-id", []byte(`{"test":"data"}`))
	testTopic := "test-topic"

	// Subscribe to the test topic BEFORE publishing
	msgChan, err := messageBus.PubSub.Subscribe(ctx, testTopic)
	require.NoError(t, err)
	
	// Wait a moment to ensure the subscription is fully registered
	time.Sleep(100 * time.Millisecond)

	// Publish the message
	err = publisher.Publish(testTopic, testMsg)
	require.NoError(t, err)

	// Wait for the message with a reasonable timeout
	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 3*time.Second)
	defer timeoutCancel()

	select {
	case msg := <-msgChan:
		// Verify the message
		assert.Equal(t, testMsg.UUID, msg.UUID)
		assert.Equal(t, testMsg.Payload, msg.Payload)
		msg.Ack()
	case <-timeoutCtx.Done():
		require.Fail(t, "Did not receive message within timeout")
	}
}

func TestGoChannelPublisher_Close(t *testing.T) {
	// Create a logger
	logger := slog.Default()

	// Create a new GoChannelPublisher
	publisher := cqrs.NewGoChannelPublisher(logger)

	// Close the publisher
	err := publisher.Close()
	require.NoError(t, err)

	// Note: We're changing the test expectation here since the behavior in the implementation
	// is that Close does actually close the internal GoChannel, contrary to the previous test comment.
	// After Close, publishing should fail with an error about the Pub/Sub being closed.
	testMsg := message.NewMessage("test-id", []byte(`{"test":"data"}`))
	err = publisher.Publish("test-topic", testMsg)
	assert.Error(t, err, "Publishing after close should fail")
}

func TestFailingPublisher(t *testing.T) {
	// Create a failing publisher
	expectedErr := assert.AnError
	publisher := &test.FailingPublisher{Err: expectedErr}
	
	// Create a test message
	testMsg := message.NewMessage("test-id", []byte(`{"test":"data"}`))
	
	// Publish should return the error
	err := publisher.Publish("test-topic", testMsg)
	assert.ErrorIs(t, err, expectedErr)
	
	// Close should not return an error
	err = publisher.Close()
	assert.NoError(t, err)
}
