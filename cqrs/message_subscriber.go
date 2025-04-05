package cqrs

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/IBM/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/redis/go-redis/v9"
)

// SubscriberConstructor is a function type that creates a message.Subscriber for a specified consumer group.
// It encapsulates the logic for creating different types of subscribers (Redis, Kafka, GoChannel)
// while allowing the consumer group to be specified at subscription time.
type SubscriberConstructor func(consumerGroup string) (message.Subscriber, error)

// NewRedisSubscriber creates a Redis-backed message subscriber constructor with default configuration.
// It requires a Redis client and a logger, and returns a function that can create subscribers for
// specific consumer groups.
func NewRedisSubscriber(redisClient redis.UniversalClient, log *slog.Logger) SubscriberConstructor {
	return NewRedisSubscriberWithConfig(redisClient, redisstream.SubscriberConfig{
		Client:       redisClient,
		Unmarshaller: redisstream.DefaultMarshallerUnmarshaller{},
	}, log)
}

// NewRedisSubscriberWithConfig creates a Redis-backed message subscriber constructor with custom configuration.
// It allows full control over the Redis subscriber configuration options.
//
// Parameters:
//   - redisClient: The Redis client to use for subscribing to messages
//   - config: Custom Redis subscriber configuration
//   - log: Logger for subscriber operations
//
// Returns a function that creates a Redis subscriber for a specific consumer group.
func NewRedisSubscriberWithConfig(
	redisClient redis.UniversalClient,
	config redisstream.SubscriberConfig,
	log *slog.Logger,
) SubscriberConstructor {
	return func(consumerGroup string) (message.Subscriber, error) {
		if consumerGroup == "" {
			return nil, fmt.Errorf("consumer group cannot be empty")
		}
		config.ConsumerGroup = consumerGroup
		return redisstream.NewSubscriber(config, watermill.NewSlogLogger(log))
	}
}

// NewGoChannelSubscriber creates an in-memory Go channel subscriber constructor with default configuration.
// This is useful for testing or for in-process message passing without external dependencies.
//
// Parameters:
//   - ctx: Context used by the subscriber
//   - log: Logger for subscriber operations
//
// Returns a function that creates a Go channel subscriber for a specific consumer group.
func NewGoChannelSubscriber(ctx context.Context, log *slog.Logger) SubscriberConstructor {
	return NewGoChannelSubscriberWithConfig(ctx, gochannel.Config{
		OutputChannelBuffer:            1000,
		Persistent:                     true,
		BlockPublishUntilSubscriberAck: false,
	}, log)
}

// NewGoChannelSubscriberWithConfig creates an in-memory Go channel subscriber constructor with custom configuration.
// It allows full control over the Go channel subscriber behavior.
//
// Parameters:
//   - ctx: Context used by the subscriber
//   - config: Custom Go channel configuration
//   - log: Logger for subscriber operations
//
// Returns a function that creates a Go channel subscriber for a specific consumer group.
func NewGoChannelSubscriberWithConfig(
	ctx context.Context,
	config gochannel.Config,
	log *slog.Logger,
) SubscriberConstructor {
	return func(consumerGroup string) (message.Subscriber, error) {
		return gochannel.NewGoChannel(config, watermill.NewSlogLogger(log)), nil
	}
}

// NewKafkaSubscriber creates a Kafka-backed message subscriber constructor with default configuration.
// It configures the Kafka subscriber to start reading from the oldest messages (offset) by default.
//
// Parameters:
//   - brokers: List of Kafka broker addresses
//   - log: Logger for subscriber operations
//
// Returns a function that creates a Kafka subscriber for a specific consumer group.
func NewKafkaSubscriber(brokers []string, log *slog.Logger) SubscriberConstructor {
	// Set the default Sarama config
	saramaSubscriberConfig := kafka.DefaultSaramaSubscriberConfig()
	// equivalent of auto.offset.reset: earliest
	saramaSubscriberConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	return NewKafkaSubscriberWithConfig(kafka.SubscriberConfig{
		Brokers:               brokers,
		Unmarshaler:           kafka.DefaultMarshaler{},
		OverwriteSaramaConfig: saramaSubscriberConfig,
	}, log)
}

// NewKafkaSubscriberWithConfig creates a Kafka-backed message subscriber constructor with custom configuration.
// It allows full control over the Kafka subscriber behavior.
//
// Parameters:
//   - config: Custom Kafka subscriber configuration
//   - log: Logger for subscriber operations
//
// Returns a function that creates a Kafka subscriber for a specific consumer group.
// The function ensures that consumer group is non-empty and sets default unmarshaler and Sarama config
// if they are not provided.
func NewKafkaSubscriberWithConfig(
	config kafka.SubscriberConfig,
	log *slog.Logger,
) SubscriberConstructor {
	return func(consumerGroup string) (message.Subscriber, error) {
		if consumerGroup == "" {
			return nil, fmt.Errorf("consumer group cannot be empty")
		}
		config.ConsumerGroup = consumerGroup
		if config.Unmarshaler == nil {
			config.Unmarshaler = kafka.DefaultMarshaler{}
		}
		if config.OverwriteSaramaConfig == nil {
			config.OverwriteSaramaConfig = kafka.DefaultSaramaSubscriberConfig()
		}
		return kafka.NewSubscriber(config, watermill.NewSlogLogger(log))
	}
}
