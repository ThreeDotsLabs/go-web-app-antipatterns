package main

import (
	"context"
)

type CommandHandler[C any] interface {
	Handle(ctx context.Context, cmd C) error
}

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

type SubscribeHandler struct{}

func NewAuthorizedSubscribeHandler(logger Logger, metricsClient MetricsClient) CommandHandler[Subscribe] {
	return loggingDecorator[Subscribe]{
		base: metricsDecorator[Subscribe]{
			base: authorizationDecorator[Subscribe]{
				base: SubscribeHandler{},
			},
			client: metricsClient,
		},
		logger: logger,
	}
}

func NewUnauthorizedSubscribeHandler(logger Logger, metricsClient MetricsClient) CommandHandler[Subscribe] {
	return loggingDecorator[Subscribe]{
		base: metricsDecorator[Subscribe]{
			base:   SubscribeHandler{},
			client: metricsClient,
		},
		logger: logger,
	}
}

func NewSubscribeHandler() CommandHandler[Subscribe] {
	return SubscribeHandler{}
}

func (h SubscribeHandler) Handle(ctx context.Context, cmd Subscribe) error {
	// Subscribe the user to the newsletter
	return nil
}
