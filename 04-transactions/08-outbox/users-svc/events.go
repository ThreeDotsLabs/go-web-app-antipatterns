package main

import (
	"context"
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/message"
)

type EventPublisher struct {
	publisher message.Publisher
}

func NewEventPublisher(natsURL string) (*EventPublisher, error) {
	logger := watermill.NewStdLogger(false, false)

	publisher, err := nats.NewPublisher(
		nats.PublisherConfig{
			URL: natsURL,
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

type PointsUsedForDiscount struct {
	UserID int `json:"user_id"`
	Points int `json:"points"`
}

func (p *EventPublisher) PublishPointsUsedForDiscount(ctx context.Context, userID int, points int) error {
	event := PointsUsedForDiscount{
		UserID: userID,
		Points: points,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	msg.SetContext(ctx)

	err = p.publisher.Publish("PointsUsedForDiscount", msg)
	if err != nil {
		return err
	}

	return nil
}
