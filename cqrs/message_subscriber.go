package cqrs

import (
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

type SubscriberConstructor func(consumerGroup string) (message.Subscriber, error)

func NewRedisSubscriber(redisClient redis.UniversalClient, log *slog.Logger) SubscriberConstructor {
	return func(consumerGroup string) (message.Subscriber, error) {
		return redisstream.NewSubscriber(
			redisstream.SubscriberConfig{
				Client:        redisClient,
				Unmarshaller:  redisstream.DefaultMarshallerUnmarshaller{},
				ConsumerGroup: consumerGroup,
			},
			watermill.NewSlogLogger(log),
		)
	}
}
