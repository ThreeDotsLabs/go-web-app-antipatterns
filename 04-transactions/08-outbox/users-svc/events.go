package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	watermillSQL "github.com/ThreeDotsLabs/watermill-sql/v3/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
)

type EventPublisher struct {
	publisher message.Publisher
}

func NewEventPublisher(db *sql.Tx) (*EventPublisher, error) {
	logger := watermill.NewStdLogger(false, false)

	publisher, err := watermillSQL.NewPublisher(
		db,
		watermillSQL.PublisherConfig{
			SchemaAdapter:        watermillSQL.DefaultPostgreSQLSchema{},
			AutoInitializeSchema: true,
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	return &EventPublisher{
		publisher: publisher,
	}, nil
}

func (p *EventPublisher) Publish(ctx context.Context, event any) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	msg.SetContext(ctx)

	name := fmt.Sprintf("%T", event)
	parts := strings.Split(name, ".")
	topic := parts[len(parts)-1]

	err = p.publisher.Publish(topic, msg)
	if err != nil {
		return err
	}

	return nil
}

func NewEventsForwarder(
	natsURL string,
	db *sql.DB,
) (*forwarder.Forwarder, error) {
	logger := watermill.NewStdLogger(false, false)

	subscriber, err := watermillSQL.NewSubscriber(
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

	publisher, err := nats.NewPublisher(
		nats.PublisherConfig{
			URL: natsURL,
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	fwd, err := forwarder.NewForwarder(
		subscriber,
		publisher,
		logger,
		forwarder.Config{
			ForwarderTopic: "forwarder",
		},
	)
	if err != nil {
		return nil, err
	}

	return fwd, nil
}
