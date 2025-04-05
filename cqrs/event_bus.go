package cqrs

import (
	"context"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/dmitrymomot/gokit/utils"
	"github.com/redis/go-redis/v9"
)

// EventBus is an interface for publishing events.
// Implementations should be safe for concurrent use by multiple goroutines.
type EventBus interface {
	Publish(ctx context.Context, event any) error
}

// NewEventBus creates a new EventBus.
// Use this event bus to publish events.
func NewEventBus(redisConn redis.UniversalClient, log *slog.Logger) (EventBus, error) {
	eventPublisher, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client:     redisConn,
		Marshaller: redisstream.DefaultMarshallerUnmarshaller{},
	}, nil)
	if err != nil {
		return nil, err
	}

	eventBus, err := cqrs.NewEventBusWithConfig(
		eventPublisher,
		cqrs.EventBusConfig{
			GeneratePublishTopic: generateEventBusPublishTopic,
			Marshaler:            marshaler,
			Logger:               watermill.NewSlogLogger(log),
		},
	)
	if err != nil {
		return nil, err
	}

	return eventBus, nil
}

func generateEventBusPublishTopic(params cqrs.GenerateEventPublishTopicParams) (string, error) {
	return utils.GetNameFromStruct(params.Event, utils.StructName), nil
}
