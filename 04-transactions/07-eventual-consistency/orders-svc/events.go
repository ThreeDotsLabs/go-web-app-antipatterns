package main

import (
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/message"
)

func NewEventsRouter(
	natsURL string,
	addDiscountHandler AddDiscountHandler,
) (*message.Router, error) {
	logger := watermill.NewStdLogger(false, false)
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, err
	}

	subscriber, err := nats.NewSubscriber(
		nats.SubscriberConfig{
			URL: natsURL,
			JetStream: nats.JetStreamConfig{
				AutoProvision: true,
			},
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	router.AddNoPublisherHandler(
		"add_discount",
		"PointsUsedForDiscount",
		subscriber,
		func(msg *message.Message) error {
			type PointsUsedForDiscount struct {
				UserID int `json:"user_id"`
				Points int `json:"points"`
			}

			var payload PointsUsedForDiscount
			err := json.Unmarshal(msg.Payload, &payload)
			if err != nil {
				return err
			}

			cmd := AddDiscount{
				UserID:   payload.UserID,
				Discount: payload.Points,
			}

			err = addDiscountHandler.Handle(msg.Context(), cmd)
			if err != nil {
				return err
			}

			return nil
		},
	)

	return router, nil
}
