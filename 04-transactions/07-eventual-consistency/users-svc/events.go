package main

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
)

type WatermillEventPublisher struct {
	publisher message.Publisher
}

func NewEventPublisher(redisAddr string) (*WatermillEventPublisher, error) {
	logger := watermill.NewStdLogger(false, false)

	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	publisher, err := redisstream.NewPublisher(
		redisstream.PublisherConfig{
			Client: client,
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	return &WatermillEventPublisher{
		publisher: publisher,
	}, nil
}

type PointsUsedForDiscount struct {
	UserID int `json:"user_id"`
	Points int `json:"points"`
}

func (p *WatermillEventPublisher) PublishPointsUsedForDiscount(ctx context.Context, userID int, points int) error {
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
