package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/dmitrymomot/gokit/sse"
	"github.com/dmitrymomot/gokit/sse/brokers/redis"
	redislib "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockXReadGroupArgs struct {
	Group    string
	Consumer string
	Streams  []string
	Count    int64
	Block    time.Duration
}

// mockRedisClient is a minimal mock of redis.UniversalClient for testing
type mockRedisClient struct {
	// Embed a nil interface to ensure we're forced to implement all methods
	redislib.UniversalClient

	// Store fake args for inspection
	addedArgs     *redislib.XAddArgs
	groupArgs     struct{ stream, group, start string }
	readGroupArgs *mockXReadGroupArgs
	trimArgs      struct {
		key    string
		maxLen int64
	}
	ackArgs struct {
		stream, group string
		ids           []string
	}
	pubArgs struct {
		channel string
		message interface{}
	}

	// Return values
	addedMsgID       string
	streamLength     int64
	errorToReturn    error
	streamMessages   []redislib.XMessage
	publishReceivers int64
}

// We only need to implement the methods actually used by StreamsBroker

// XAdd implements redis client interface
func (m *mockRedisClient) XAdd(ctx context.Context, args *redislib.XAddArgs) *redislib.StringCmd {
	m.addedArgs = args
	return redislib.NewStringResult(m.addedMsgID, m.errorToReturn)
}

// XGroupCreate implements redis client interface
func (m *mockRedisClient) XGroupCreate(ctx context.Context, stream, group, start string) *redislib.StatusCmd {
	m.groupArgs.stream = stream
	m.groupArgs.group = group
	m.groupArgs.start = start

	if m.errorToReturn != nil && m.errorToReturn.Error() == "BUSYGROUP Consumer Group name already exists" {
		// This is a special case we need to test
		return redislib.NewStatusResult("", m.errorToReturn)
	}
	return redislib.NewStatusResult("OK", m.errorToReturn)
}

// XInfoStream implements redis client interface
func (m *mockRedisClient) XInfoStream(ctx context.Context, stream string) *redislib.XInfoStreamCmd {
	// This is a special case - we need to return redis.Nil if we want to test stream not existing
	if m.errorToReturn == redislib.Nil {
		cmd := redislib.NewXInfoStreamCmd(ctx, stream)
		cmd.SetErr(redislib.Nil)
		return cmd
	}

	// Normal case - return the stream info or configured error
	cmd := redislib.NewXInfoStreamCmd(ctx, stream)
	if m.errorToReturn != nil {
		cmd.SetErr(m.errorToReturn)
	} else {
		// Pass pointer to XInfoStream literal for SetVal
		cmd.SetVal(&redislib.XInfoStream{Length: m.streamLength})
	}
	return cmd
}

// XReadGroup implements redis client interface
func (m *mockRedisClient) XReadGroup(ctx context.Context, args *redislib.XReadGroupArgs) *redislib.XStreamSliceCmd {
	// Store args for inspection in tests
	m.readGroupArgs = &mockXReadGroupArgs{
		Group:    args.Group,
		Consumer: args.Consumer,
		Streams:  args.Streams,
		Count:    args.Count,
		Block:    args.Block,
	}

	// Return configured messages or error
	if m.errorToReturn == redislib.Nil {
		return redislib.NewXStreamSliceCmd(ctx, func(c context.Context, cmd redislib.Cmder) error {
			return redislib.Nil
		})
	}

	return redislib.NewXStreamSliceCmd(ctx, func(c context.Context, cmd redislib.Cmder) error {
		if len(m.streamMessages) > 0 {
			cmd.(*redislib.XStreamSliceCmd).SetVal([]redislib.XStream{
				{
					Stream:   args.Streams[0],
					Messages: m.streamMessages,
				},
			})
		} else {
			cmd.(*redislib.XStreamSliceCmd).SetVal([]redislib.XStream{})
		}
		return m.errorToReturn
	})
}

// XAck implements redis client interface
func (m *mockRedisClient) XAck(ctx context.Context, stream, group string, ids ...string) *redislib.IntCmd {
	m.ackArgs.stream = stream
	m.ackArgs.group = group
	m.ackArgs.ids = ids
	return redislib.NewIntResult(int64(len(ids)), m.errorToReturn)
}

// XTrimMaxLen implements redis client interface
func (m *mockRedisClient) XTrimMaxLen(ctx context.Context, key string, maxLen int64) *redislib.IntCmd {
	m.trimArgs.key = key
	m.trimArgs.maxLen = maxLen
	return redislib.NewIntResult(10, m.errorToReturn) // Return 10 as number of trimmed entries
}

// Publish implements redis client interface (for legacy broker compatibility)
func (m *mockRedisClient) Publish(ctx context.Context, channel string, message interface{}) *redislib.IntCmd {
	m.pubArgs.channel = channel
	m.pubArgs.message = message
	return redislib.NewIntResult(m.publishReceivers, m.errorToReturn)
}

// Close implements redis client interface
func (m *mockRedisClient) Close() error {
	return m.errorToReturn
}

// Helper function to create a default mock
func newMockRedisClient() *mockRedisClient {
	return &mockRedisClient{
		addedMsgID:       "mock-msg-id-1",
		streamLength:     0,
		errorToReturn:    nil,
		publishReceivers: 1,
	}
}

// TestNewStreamsBroker tests the creation of a new StreamsBroker
func TestNewStreamsBroker(t *testing.T) {
	t.Run("returns error when client is nil", func(t *testing.T) {
		broker, err := redis.NewStreamsBroker(nil)
		require.Error(t, err)
		require.ErrorIs(t, err, redis.ErrNoRedisClient)
		require.Nil(t, broker)
	})

	t.Run("creates broker with default options", func(t *testing.T) {
		client := newMockRedisClient()
		broker, err := redis.NewStreamsBroker(client)
		require.NoError(t, err)
		require.NotNil(t, broker)

		// Cleanup
		err = broker.Close()
		require.NoError(t, err)
	})

	t.Run("creates broker with custom options", func(t *testing.T) {
		client := newMockRedisClient()
		broker, err := redis.NewStreamsBroker(client, redis.StreamsOptions{
			StreamName:       "custom_stream",
			MaxStreamLength:  500,
			MessageRetention: 12 * time.Hour,
			GroupName:        "custom_group",
			ConsumerName:     "custom_consumer",
			BlockDuration:    200 * time.Millisecond,
		})
		require.NoError(t, err)
		require.NotNil(t, broker)

		// Verify stream name was used
		assert.Equal(t, "custom_stream", client.groupArgs.stream)

		// Cleanup
		err = broker.Close()
		require.NoError(t, err)
	})
}

// TestStreamsBrokerPublish tests the publish functionality
func TestStreamsBrokerPublish(t *testing.T) {
	t.Run("skips expired messages", func(t *testing.T) {
		client := newMockRedisClient()
		broker, err := redis.NewStreamsBroker(client)
		require.NoError(t, err)

		// Create a message that's already expired
		expiredMsg := sse.Message{
			Event:     "test",
			Data:      "test data",
			Timestamp: time.Now().Add(-2 * time.Minute),
			TTL:       time.Minute, // Expired 1 minute ago
		}

		// Publish should not error and should skip the message
		err = broker.Publish(context.Background(), expiredMsg)
		require.NoError(t, err)

		// Verify no message was actually sent (args should be nil)
		assert.Nil(t, client.addedArgs)

		// Cleanup
		err = broker.Close()
		require.NoError(t, err)
	})

	t.Run("publishes valid message", func(t *testing.T) {
		client := newMockRedisClient()
		broker, err := redis.NewStreamsBroker(client)
		require.NoError(t, err)

		// Create a valid message
		validMsg := sse.Message{
			Event:     "test",
			Data:      "test data",
			Timestamp: time.Now(),
		}

		// Publish should not error
		err = broker.Publish(context.Background(), validMsg)
		require.NoError(t, err)

		// Verify message was added to stream
		assert.NotNil(t, client.addedArgs)
		assert.Equal(t, "sse_messages_stream", client.addedArgs.Stream)
		assert.Equal(t, "*", client.addedArgs.ID)
		assert.Contains(t, client.addedArgs.Values, "message")

		// Cleanup
		err = broker.Close()
		require.NoError(t, err)
	})

	t.Run("handles validation error", func(t *testing.T) {
		client := newMockRedisClient()
		broker, err := redis.NewStreamsBroker(client)
		require.NoError(t, err)

		// Create an invalid message (no event)
		invalidMsg := sse.Message{
			Data:      "test data",
			Timestamp: time.Now(),
		}

		// Publish should return an error
		err = broker.Publish(context.Background(), invalidMsg)
		require.Error(t, err)

		// Cleanup
		err = broker.Close()
		require.NoError(t, err)
	})
}

// TestStreamsBrokerSubscribe tests the subscribe functionality
func TestStreamsBrokerSubscribe(t *testing.T) {
	t.Run("subscribes successfully", func(t *testing.T) {
		client := newMockRedisClient()
		broker, err := redis.NewStreamsBroker(client)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// Subscribe should not error
		ch, err := broker.Subscribe(ctx)
		require.NoError(t, err)
		require.NotNil(t, ch)

		// Verify channel is closed when context is cancelled
		<-ctx.Done()
		_, ok := <-ch
		assert.False(t, ok, "channel should be closed")

		// Cleanup
		err = broker.Close()
		require.NoError(t, err)
	})

	t.Run("fails when broker is closed", func(t *testing.T) {
		client := newMockRedisClient()
		broker, err := redis.NewStreamsBroker(client)
		require.NoError(t, err)

		// Close the broker
		err = broker.Close()
		require.NoError(t, err)

		// Subscribe should error
		ch, err := broker.Subscribe(context.Background())
		require.Error(t, err)
		require.ErrorIs(t, err, sse.ErrBrokerClosed)
		require.Nil(t, ch)
	})
}

// TestStreamsBrokerCleanupStream tests the cleanup functionality
func TestStreamsBrokerCleanupStream(t *testing.T) {
	t.Run("cleans up stream", func(t *testing.T) {
		client := newMockRedisClient()
		broker, err := redis.NewStreamsBroker(client)
		require.NoError(t, err)

		// Cleanup should not error
		err = broker.CleanupStream(context.Background())
		require.NoError(t, err)

		// Verify trim was called with correct parameters
		assert.Equal(t, "sse_messages_stream", client.trimArgs.key)
		assert.Equal(t, int64(1000), client.trimArgs.maxLen)

		// Cleanup
		err = broker.Close()
		require.NoError(t, err)
	})

	t.Run("skips cleanup when max length is 0", func(t *testing.T) {
		client := newMockRedisClient()
		broker, err := redis.NewStreamsBroker(client, redis.StreamsOptions{
			MaxStreamLength: 0, // No max length
		})
		require.NoError(t, err)

		broker.SetMaxStreamLength(0)

		// Record initial state
		initialTrimKey := client.trimArgs.key

		// Cleanup should not error
		err = broker.CleanupStream(context.Background())
		require.NoError(t, err)

		// Verify trim was not called (args unchanged)
		assert.Equal(t, initialTrimKey, client.trimArgs.key)

		// Cleanup
		err = broker.Close()
		require.NoError(t, err)
	})
}
