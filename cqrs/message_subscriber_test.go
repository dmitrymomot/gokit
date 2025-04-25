package cqrs_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/dmitrymomot/gokit/cqrs"
	"github.com/dmitrymomot/gokit/cqrs/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGoChannelSubscriber(t *testing.T) {
	// Setup
	ctx := context.Background()
	logger := slog.Default()

	// Test
	constructor := cqrs.NewGoChannelSubscriber(ctx, logger)

	// Assert
	require.NotNil(t, constructor)

	// Test creating a subscriber with the constructor
	subscriber, err := constructor("test-consumer-group")
	require.NoError(t, err)
	require.NotNil(t, subscriber)
}

func TestGoChannelSubscriberWithConfig(t *testing.T) {
	// Setup
	ctx := context.Background()
	logger := slog.Default()

	// Test with custom config
	config := gochannel.Config{
		OutputChannelBuffer: 100,
		Persistent:          true,
	}

	constructor := cqrs.NewGoChannelSubscriberWithConfig(
		ctx,
		config,
		logger,
	)

	// Assert
	require.NotNil(t, constructor)

	// Test creating a subscriber with the constructor
	subscriber, err := constructor("test-consumer-group")
	require.NoError(t, err)
	require.NotNil(t, subscriber)
}

func TestGoChannelSubscriberConstructor_EmptyConsumerGroup(t *testing.T) {
	// Setup
	ctx := context.Background()
	logger := slog.Default()

	// Test
	constructor := cqrs.NewGoChannelSubscriber(ctx, logger)

	// Assert - Empty consumer group should still work with GoChannel
	subscriber, err := constructor("")
	require.NoError(t, err)
	require.NotNil(t, subscriber)
}

func TestGoChannelSubscriber_SubscribeAndPublish(t *testing.T) {
	// Setup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger := slog.Default()

	// Create a shared message bus for both publisher and subscriber
	config := test.NewDefaultTestBusConfig()
	config.Logger = logger
	messageBus, _ := test.NewChannelMessageBus(t, ctx, config)

	// Subscribe to a topic
	const testTopic = "test-topic"
	msgChan, err := messageBus.PubSub.Subscribe(ctx, testTopic)
	require.NoError(t, err)

	// Wait a moment for subscription to be fully registered
	time.Sleep(100 * time.Millisecond)

	// Publish a message
	testMessage := message.NewMessage(
		"test-id",
		[]byte(`{"name":"test","value":42}`),
	)

	err = messageBus.PubSub.Publish(testTopic, testMessage)
	require.NoError(t, err)

	// Wait for the message with a shorter timeout
	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 2*time.Second)
	defer timeoutCancel()

	select {
	case msg := <-msgChan:
		// Verify message
		assert.Equal(t, testMessage.UUID, msg.UUID)
		assert.Equal(t, testMessage.Payload, msg.Payload)

		// Acknowledge message
		msg.Ack()
	case <-timeoutCtx.Done():
		require.Fail(t, "Timeout waiting for message")
	}
}

func TestGoChannelSubscriber_SubscribeMultipleTopics(t *testing.T) {
	// Setup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger := slog.Default()

	// Create a shared message bus for publisher and subscriber
	config := test.NewDefaultTestBusConfig()
	config.Logger = logger
	messageBus, _ := test.NewChannelMessageBus(t, ctx, config)

	// Subscribe to multiple topics
	const topic1 = "topic-1"
	const topic2 = "topic-2"

	msgChan1, err := messageBus.PubSub.Subscribe(ctx, topic1)
	require.NoError(t, err)

	msgChan2, err := messageBus.PubSub.Subscribe(ctx, topic2)
	require.NoError(t, err)

	// Wait a moment for subscriptions to be fully registered
	time.Sleep(100 * time.Millisecond)

	// Publish messages to both topics
	msg1 := message.NewMessage("id-1", []byte("payload-1"))
	msg2 := message.NewMessage("id-2", []byte("payload-2"))

	err = messageBus.PubSub.Publish(topic1, msg1)
	require.NoError(t, err)

	err = messageBus.PubSub.Publish(topic2, msg2)
	require.NoError(t, err)

	// Wait for messages from topic1 with shorter timeout
	timeoutCtx1, cancel1 := context.WithTimeout(ctx, 2*time.Second)
	defer cancel1()

	select {
	case receivedMsg := <-msgChan1:
		assert.Equal(t, msg1.UUID, receivedMsg.UUID)
		assert.Equal(t, msg1.Payload, receivedMsg.Payload)
		receivedMsg.Ack()
	case <-timeoutCtx1.Done():
		require.Fail(t, "Timeout waiting for message from topic1")
	}

	// Wait for messages from topic2
	timeoutCtx2, cancel2 := context.WithTimeout(ctx, 2*time.Second)
	defer cancel2()

	select {
	case receivedMsg := <-msgChan2:
		assert.Equal(t, msg2.UUID, receivedMsg.UUID)
		assert.Equal(t, msg2.Payload, receivedMsg.Payload)
		receivedMsg.Ack()
	case <-timeoutCtx2.Done():
		require.Fail(t, "Timeout waiting for message from topic2")
	}
}

func TestGoChannelSubscriber_Close(t *testing.T) {
	// Setup
	ctx := context.Background()
	logger := slog.Default()

	// Create subscriber
	subscriberConstructor := cqrs.NewGoChannelSubscriber(ctx, logger)
	subscriber, err := subscriberConstructor("test-consumer")
	require.NoError(t, err)

	// Test closing the subscriber
	err = subscriber.Close()
	assert.NoError(t, err)
}

func TestMultipleSubscribersPerTopic(t *testing.T) {
	// Setup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger := slog.Default()

	// Create a shared message bus
	config := test.NewDefaultTestBusConfig()
	config.Logger = logger
	messageBus, _ := test.NewChannelMessageBus(t, ctx, config)

	// Subscribe to the same topic with two different subscriber instances
	const testTopic = "test-topic"

	// Create two separate subscribers
	subscriber1 := messageBus.PubSub
	subscriber2 := messageBus.PubSub

	// Subscribe with different consumer groups
	msgChan1, err := subscriber1.Subscribe(ctx, testTopic)
	require.NoError(t, err)

	msgChan2, err := subscriber2.Subscribe(ctx, testTopic)
	require.NoError(t, err)

	// Wait a moment for subscriptions to be fully registered
	time.Sleep(100 * time.Millisecond)

	// Publish a message
	testMessage := message.NewMessage(
		"test-id",
		[]byte(`{"name":"test","value":42}`),
	)

	err = messageBus.PubSub.Publish(testTopic, testMessage)
	require.NoError(t, err)

	// Both subscribers should receive the message
	// Wait for subscriber 1 with shorter timeout
	timeoutCtx1, cancel1 := context.WithTimeout(ctx, 2*time.Second)
	defer cancel1()

	select {
	case msg := <-msgChan1:
		assert.Equal(t, testMessage.UUID, msg.UUID)
		assert.Equal(t, testMessage.Payload, msg.Payload)
		msg.Ack()
	case <-timeoutCtx1.Done():
		require.Fail(t, "Timeout waiting for message on subscriber 1")
	}

	// Wait for subscriber 2
	timeoutCtx2, cancel2 := context.WithTimeout(ctx, 2*time.Second)
	defer cancel2()

	select {
	case msg := <-msgChan2:
		assert.Equal(t, testMessage.UUID, msg.UUID)
		assert.Equal(t, testMessage.Payload, msg.Payload)
		msg.Ack()
	case <-timeoutCtx2.Done():
		require.Fail(t, "Timeout waiting for message on subscriber 2")
	}
}
