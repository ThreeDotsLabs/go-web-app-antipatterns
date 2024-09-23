package main

import (
	"context"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

type PointsUsedForDiscount struct {
	UserID int `json:"user_id"`
	Points int `json:"points"`
}

func (p *PointsUsedForDiscount) EventName() string {
	return "PointsUsedForDiscount"
}

func NewEventsRouter(
	redisAddr string,
	addDiscountHandler AddDiscountHandler,
) (*message.Router, error) {
	logger := watermill.NewStdLogger(false, false)
	router := message.NewDefaultRouter(logger)

	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	eventProcessor, err := cqrs.NewEventProcessorWithConfig(router, cqrs.EventProcessorConfig{
		GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
			return params.EventName, nil
		},
		SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
			return redisstream.NewSubscriber(
				redisstream.SubscriberConfig{
					Client:        client,
					ConsumerGroup: "orders-svc",
				},
				logger,
			)
		},
		Marshaler: cqrs.JSONMarshaler{
			GenerateName: cqrs.EventName,
		},
		Logger: logger,
	})
	if err != nil {
		return nil, err
	}

	err = eventProcessor.AddHandlers(
		cqrs.NewEventHandler(
			"add-discount",
			func(ctx context.Context, event *PointsUsedForDiscount) error {
				cmd := AddDiscount{
					UserID:   event.UserID,
					Discount: event.Points,
				}

				err = addDiscountHandler.Handle(ctx, cmd)
				if err != nil {
					return err
				}

				return nil
			},
		),
	)
	if err != nil {
		return nil, err
	}

	return router, nil
}
