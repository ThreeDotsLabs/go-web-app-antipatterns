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

type OnPointsUsedForDiscountHandler struct {
	addDiscountHandler AddDiscountHandler
}

func (h OnPointsUsedForDiscountHandler) Handle(ctx context.Context, event *PointsUsedForDiscount) error {
	cmd := AddDiscount{
		UserID:   event.UserID,
		Discount: event.Points,
	}

	return h.addDiscountHandler.Handle(ctx, cmd)
}

func NewEventsRouter(
	redisAddr string,
	addDiscountHandler AddDiscountHandler,
) (*message.Router, error) {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	logger := watermill.NewStdLogger(false, false)

	router := message.NewDefaultRouter(logger)

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
		Marshaler: cqrs.JSONMarshaler{},
		Logger:    logger,
	})
	if err != nil {
		return nil, err
	}

	onPointsUsedForDiscountHandler := OnPointsUsedForDiscountHandler{
		addDiscountHandler: addDiscountHandler,
	}

	err = eventProcessor.AddHandlers(
		cqrs.NewEventHandler(
			"OnPointsUsedForDiscountHandler",
			onPointsUsedForDiscountHandler.Handle,
		),
	)
	if err != nil {
		return nil, err
	}

	return router, nil
}
