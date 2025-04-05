package cqrs

import (
	"context"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/dmitrymomot/gokit/utils"
)

// CommandBus is an interface for sending commands.
// Implementations should be safe for concurrent use by multiple goroutines.
type CommandBus interface {
	Send(ctx context.Context, cmd any) error
	SendWithModifiedMessage(ctx context.Context, cmd any, modify func(*message.Message) error) error
}

// NewCommandBus creates a new CommandBus.
// Use this command bus to send commands to the service.
func NewCommandBus(publisher message.Publisher, log *slog.Logger) (CommandBus, error) {
	commandBus, err := cqrs.NewCommandBusWithConfig(
		publisher,
		cqrs.CommandBusConfig{
			GeneratePublishTopic: generateCommandBusPublishTopic,
			Marshaler:            marshaler,
			Logger:               watermill.NewSlogLogger(log),
		},
	)
	if err != nil {
		return nil, err
	}

	return commandBus, nil
}

func generateCommandBusPublishTopic(params cqrs.CommandBusGeneratePublishTopicParams) (string, error) {
	return utils.GetNameFromStruct(params.Command, utils.StructName), nil
}
