package main

import (
	"context"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/redis/go-redis/v9"
)

type WatermillEventPublisher struct {
	eventBus *cqrs.EventBus
}

func NewEventPublisher(redisAddr string) (*WatermillEventPublisher, error) {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	logger := watermill.NewStdLogger(false, false)

	publisher, err := redisstream.NewPublisher(
		redisstream.PublisherConfig{
			Client: client,
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

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

type PointsUsedForDiscount struct {
	UserID int `json:"user_id"`
	Points int `json:"points"`
}

func (p PointsUsedForDiscount) EventName() string {
	return "PointsUsedForDiscount"
}

func (p *WatermillEventPublisher) PublishPointsUsedForDiscount(ctx context.Context, userID int, points int) error {
	event := PointsUsedForDiscount{
		UserID: userID,
		Points: points,
	}

	err := p.eventBus.Publish(ctx, event)
	if err != nil {
		return err
	}

	return nil
}
