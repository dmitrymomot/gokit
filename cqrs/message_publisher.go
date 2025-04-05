package cqrs

import (
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/redis/go-redis/v9"
)

// NewRedisPublisher creates a new Redis stream publisher with default configuration.
// It uses the provided Redis client connection and logger.
//
// Parameters:
//   - redisConn: A Redis client connection implementing the UniversalClient interface
//   - log: A structured logger from the slog package
//
// Returns:
//   - message.Publisher: A Watermill message publisher interface
//   - error: Any error encountered during publisher creation
func NewRedisPublisher(redisConn redis.UniversalClient, log *slog.Logger) (message.Publisher, error) {
	publisher, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client:     redisConn,
		Marshaller: redisstream.DefaultMarshallerUnmarshaller{},
	}, watermill.NewSlogLogger(log))
	if err != nil {
		return nil, err
	}

	return publisher, nil
}

// NewRedisPublisherWithConfig creates a new Redis stream publisher with custom configuration.
// It allows full control over the Redis publisher settings.
//
// Parameters:
//   - redisConn: A Redis client connection implementing the UniversalClient interface
//   - config: Custom configuration for the Redis publisher
//   - log: A structured logger from the slog package
//
// Returns:
//   - message.Publisher: A Watermill message publisher interface
//   - error: Any error encountered during publisher creation
func NewRedisPublisherWithConfig(redisConn redis.UniversalClient, config redisstream.PublisherConfig, log *slog.Logger) (message.Publisher, error) {
	publisher, err := redisstream.NewPublisher(config, watermill.NewSlogLogger(log))
	if err != nil {
		return nil, err
	}

	return publisher, nil
}

// NewKafkaPublisher creates a new Kafka publisher with default configuration.
// It connects to the specified Kafka brokers and uses the default marshaler.
//
// Parameters:
//   - brokers: A slice of strings containing Kafka broker addresses
//   - log: A structured logger from the slog package
//
// Returns:
//   - message.Publisher: A Watermill message publisher interface
//   - error: Any error encountered during publisher creation
func NewKafkaPublisher(brokers []string, log *slog.Logger) (message.Publisher, error) {
	publisher, err := kafka.NewPublisher(kafka.PublisherConfig{
		Brokers:   brokers,
		Marshaler: kafka.DefaultMarshaler{},
	}, watermill.NewSlogLogger(log))
	if err != nil {
		return nil, err
	}

	return publisher, nil
}

// NewKafkaPublisherWithConfig creates a new Kafka publisher with custom configuration.
// It allows full control over the Kafka publisher settings.
//
// Parameters:
//   - brokers: A slice of strings containing Kafka broker addresses
//   - config: Custom configuration for the Kafka publisher
//   - log: A structured logger from the slog package
//
// Returns:
//   - message.Publisher: A Watermill message publisher interface
//   - error: Any error encountered during publisher creation
func NewKafkaPublisherWithConfig(brokers []string, config kafka.PublisherConfig, log *slog.Logger) (message.Publisher, error) {
	publisher, err := kafka.NewPublisher(config, watermill.NewSlogLogger(log))
	if err != nil {
		return nil, err
	}

	return publisher, nil
}

// NewGoChannelPublisher creates a new in-memory publisher using Go channels with default configuration.
// This publisher is particularly useful for testing or single-process applications.
//
// Parameters:
//   - log: A structured logger from the slog package
//
// Returns:
//   - message.Publisher: A Watermill message publisher interface
func NewGoChannelPublisher(log *slog.Logger) message.Publisher {
	return gochannel.NewGoChannel(gochannel.Config{
		OutputChannelBuffer:            1000,
		Persistent:                     true,
		BlockPublishUntilSubscriberAck: false,
	}, watermill.NewSlogLogger(log))
}

// NewGoChannelPublisherWithConfig creates a new in-memory publisher using Go channels with custom configuration.
// It allows full control over the Go channel publisher settings.
//
// Parameters:
//   - config: Custom configuration for the Go channel publisher
//   - log: A structured logger from the slog package
//
// Returns:
//   - message.Publisher: A Watermill message publisher interface
func NewGoChannelPublisherWithConfig(config gochannel.Config, log *slog.Logger) message.Publisher {
	return gochannel.NewGoChannel(config, watermill.NewSlogLogger(log))
}
