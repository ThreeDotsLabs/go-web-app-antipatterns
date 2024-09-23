package main

import (
	"context"
	"database/sql"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	watermillSQL "github.com/ThreeDotsLabs/watermill-sql/v3/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

const forwarderTopic = "forwarder"

type WatermillEventPublisher struct {
	eventBus *cqrs.EventBus
}

func NewEventPublisher(db *sql.Tx) (*WatermillEventPublisher, error) {
	logger := watermill.NewStdLogger(false, false)

	var publisher message.Publisher
	var err error

	publisher, err = watermillSQL.NewPublisher(
		db,
		watermillSQL.PublisherConfig{
			SchemaAdapter: watermillSQL.DefaultPostgreSQLSchema{},
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	publisher = forwarder.NewPublisher(
		publisher,
		forwarder.PublisherConfig{
			ForwarderTopic: forwarderTopic,
		},
	)

	eventBus, err := cqrs.NewEventBusWithConfig(publisher, cqrs.EventBusConfig{
		GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
			return params.EventName, nil
		},
		Marshaler: cqrs.JSONMarshaler{
			GenerateName: cqrs.EventName,
		},
		Logger: logger,
	})
	if err != nil {
		return nil, err
	}

	return &WatermillEventPublisher{
		eventBus: eventBus,
	}, nil
}

func (p *WatermillEventPublisher) Publish(ctx context.Context, event cqrs.Event) error {
	err := p.eventBus.Publish(ctx, event)
	if err != nil {
		return err
	}

	return nil
}

func NewEventsForwarder(
	redisAddr string,
	db *sql.DB,
) (*forwarder.Forwarder, error) {
	logger := watermill.NewStdLogger(false, false)

	sqlSubscriber, err := watermillSQL.NewSubscriber(
		db,
		watermillSQL.SubscriberConfig{
			SchemaAdapter:    watermillSQL.DefaultPostgreSQLSchema{},
			OffsetsAdapter:   watermillSQL.DefaultPostgreSQLOffsetsAdapter{},
			InitializeSchema: true,
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	redisPublisher, err := redisstream.NewPublisher(
		redisstream.PublisherConfig{
			Client: client,
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	fwd, err := forwarder.NewForwarder(
		sqlSubscriber,
		redisPublisher,
		logger,
		forwarder.Config{
			ForwarderTopic: forwarderTopic,
		},
	)
	if err != nil {
		return nil, err
	}

	return fwd, nil
}
