package main

import (
	"context"
	"errors"
	"time"
)

type MetricsClient interface {
	Inc(key string, value int)
}

type Logger interface {
	Println(args ...interface{})
}

type Subscribe struct {
	Email        string
	NewsletterID string
}

type SubscribeHandler struct {
	logger        Logger
	metricsClient MetricsClient
}

func NewSubscribeHandler(logger Logger, metricsClient MetricsClient) SubscribeHandler {
	return SubscribeHandler{
		logger:        logger,
		metricsClient: metricsClient,
	}
}

func (h SubscribeHandler) Handle(ctx context.Context, cmd Subscribe) (err error) {
	start := time.Now()
	h.logger.Println("Subscribing to newsletter", cmd)
	defer func() {
		end := time.Since(start)
		h.metricsClient.Inc("commands.subscribe.duration", int(end.Seconds()))

		if err == nil {
			h.metricsClient.Inc("commands.subscribe.success", 1)
			h.logger.Println("Subscribed to newsletter")
		} else {
			h.metricsClient.Inc("commands.subscribe.failure", 1)
			h.logger.Println("Failed subscribing to newsletter:", err)
		}
	}()

	user, err := UserFromContext(ctx)
	if err != nil {
		return err
	}

	if !user.Active {
		return errors.New("the user's account is not active")
	}

	// Subscribe the user to the newsletter
	return nil
}
