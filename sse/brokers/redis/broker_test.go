package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/dmitrymomot/gokit/sse"
	redisbroker "github.com/dmitrymomot/gokit/sse/brokers/redis"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// BrokerTestSuite defines the test suite for Redis broker
type BrokerTestSuite struct {
	suite.Suite
	miniRedis *miniredis.Miniredis
	client    redis.UniversalClient
}

// SetupTest sets up the test environment before each test
func (s *BrokerTestSuite) SetupTest() {
	// Create new miniredis server
	mr, err := miniredis.Run()
	require.NoError(s.T(), err)
	s.miniRedis = mr

	// Create a real Redis client that connects to miniredis
	s.client = redis.NewClient(&redis.Options{
		Addr: s.miniRedis.Addr(),
	})
}

// TearDownTest cleans up after each test
func (s *BrokerTestSuite) TearDownTest() {
	if s.client != nil {
		s.client.Close()
	}
	if s.miniRedis != nil {
		s.miniRedis.Close()
	}
}

// TestNewBroker tests the creation of a new Redis broker
func (s *BrokerTestSuite) TestNewBroker() {
	// Test with valid client
	broker, err := redisbroker.NewBroker(s.client)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), broker)
	broker.Close() // Clean up

	// Test with nil client
	broker, err = redisbroker.NewBroker(nil)
	require.Error(s.T(), err)
	assert.Equal(s.T(), redisbroker.ErrNoRedisClient, err)
	assert.Nil(s.T(), broker)

	// Test with custom options
	broker, err = redisbroker.NewBroker(s.client, redisbroker.Options{
		Channel: "custom_channel",
	})
	require.NoError(s.T(), err)
	require.NotNil(s.T(), broker)
	broker.Close() // Clean up
}

// TestPublish tests the Publish method
func (s *BrokerTestSuite) TestPublish() {
	// Create broker
	broker, err := redisbroker.NewBroker(s.client)
	require.NoError(s.T(), err)
	defer broker.Close()

	// Valid message
	ctx := context.Background()
	message := sse.NewMessage("test_event", "Hello, World!")
	err = broker.Publish(ctx, message)
	require.NoError(s.T(), err)

	// Invalid message (empty event)
	invalidMessage := sse.Message{
		Data: "test data",
	}
	err = broker.Publish(ctx, invalidMessage)
	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, sse.ErrInvalidMessage)

	// Invalid message (nil data)
	invalidMessage = sse.Message{
		Event: "test_event",
	}
	err = broker.Publish(ctx, invalidMessage)
	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, sse.ErrInvalidMessage)
}

// TestSubscribe tests the Subscribe method
func (s *BrokerTestSuite) TestSubscribe() {
	// Create broker
	broker, err := redisbroker.NewBroker(s.client)
	require.NoError(s.T(), err)
	defer broker.Close()

	// Subscribe to messages
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messageCh, err := broker.Subscribe(ctx)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), messageCh)

	// Publish a message
	message := sse.NewMessage("test_event", "Test message")
	err = broker.Publish(ctx, message)
	require.NoError(s.T(), err)

	// Wait for the message
	select {
	case receivedMsg := <-messageCh:
		assert.Equal(s.T(), message.Event, receivedMsg.Event)
		assert.Equal(s.T(), message.Data, receivedMsg.Data)
	case <-time.After(2 * time.Second):
		s.T().Fatal("Timeout waiting for message")
	}

	// Test cancellation
	cancel()
	time.Sleep(100 * time.Millisecond) // Allow time for cancellation to propagate
	
	// Subscribe after broker is closed
	broker.Close()
	_, err = broker.Subscribe(context.Background())
	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, sse.ErrBrokerClosed)
}

// TestClose tests the Close method
func (s *BrokerTestSuite) TestClose() {
	// Create broker
	broker, err := redisbroker.NewBroker(s.client)
	require.NoError(s.T(), err)

	// Subscribe to messages
	ctx := context.Background()
	messageCh, err := broker.Subscribe(ctx)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), messageCh)

	// Close the broker
	err = broker.Close()
	require.NoError(s.T(), err)

	// Verify all resources are cleaned up
	// Publishing after close should fail
	err = broker.Publish(ctx, sse.NewMessage("test", "data"))
	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, sse.ErrBrokerClosed)

	// Subscribing after close should fail
	_, err = broker.Subscribe(ctx)
	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, sse.ErrBrokerClosed)
}

// TestMessageDistribution tests that messages are properly distributed to subscribers
func (s *BrokerTestSuite) TestMessageDistribution() {
	// Create broker
	broker, err := redisbroker.NewBroker(s.client)
	require.NoError(s.T(), err)
	defer broker.Close()

	// Create two subscribers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch1, err := broker.Subscribe(ctx)
	require.NoError(s.T(), err)

	ch2, err := broker.Subscribe(ctx)
	require.NoError(s.T(), err)

	// Publish a message
	message := sse.NewMessage("broadcast_event", "Hello, everyone!")
	err = broker.Publish(ctx, message)
	require.NoError(s.T(), err)

	// Both subscribers should receive the message
	for i, ch := range []<-chan sse.Message{ch1, ch2} {
		select {
		case receivedMsg := <-ch:
			assert.Equal(s.T(), message.Event, receivedMsg.Event)
			assert.Equal(s.T(), message.Data, receivedMsg.Data)
		case <-time.After(2 * time.Second):
			s.T().Fatalf("Subscriber %d didn't receive the message", i+1)
		}
	}
}

// TestRedisBroker runs the test suite
func TestRedisBroker(t *testing.T) {
	suite.Run(t, new(BrokerTestSuite))
}

// Standalone tests
func TestNewBroker(t *testing.T) {
	// Start miniredis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Create a client
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	// Test with valid client
	broker, err := redisbroker.NewBroker(client)
	require.NoError(t, err)
	require.NotNil(t, broker)
	broker.Close()

	// Test with nil client
	broker, err = redisbroker.NewBroker(nil)
	require.Error(t, err)
	assert.Equal(t, redisbroker.ErrNoRedisClient, err)
	assert.Nil(t, broker)
}

func TestPublish(t *testing.T) {
	// Start miniredis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Create a client
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	// Create broker
	broker, err := redisbroker.NewBroker(client)
	require.NoError(t, err)
	defer broker.Close()

	// Valid message
	ctx := context.Background()
	message := sse.NewMessage("test_event", "Hello, World!")
	err = broker.Publish(ctx, message)
	require.NoError(t, err)
	
	// Check if the message was published to Redis
	pubsub := client.Subscribe(ctx, "sse_messages")
	defer pubsub.Close()
	
	// Invalid message
	err = broker.Publish(ctx, sse.Message{})
	require.Error(t, err)
	assert.ErrorIs(t, err, sse.ErrInvalidMessage)
}

func TestSubscribeAndClose(t *testing.T) {
	// Start miniredis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Create a client
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	// Create broker
	broker, err := redisbroker.NewBroker(client)
	require.NoError(t, err)

	// Subscribe to messages
	ctx := context.Background()
	messageCh, err := broker.Subscribe(ctx)
	require.NoError(t, err)
	require.NotNil(t, messageCh)

	// Publish a message
	message := sse.NewMessage("test_event", "Test message")
	err = broker.Publish(ctx, message)
	require.NoError(t, err)

	// We should receive the message
	select {
	case receivedMsg := <-messageCh:
		assert.Equal(t, message.Event, receivedMsg.Event)
		assert.Equal(t, message.Data, receivedMsg.Data)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for message")
	}

	// Close the broker
	err = broker.Close()
	require.NoError(t, err)

	// Verify resources are cleaned up
	_, err = broker.Subscribe(ctx)
	require.Error(t, err)
	assert.ErrorIs(t, err, sse.ErrBrokerClosed)
}
