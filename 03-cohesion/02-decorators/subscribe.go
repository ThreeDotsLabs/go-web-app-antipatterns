package main

import (
	"context"
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

type SubscribeHandler interface {
	Handle(ctx context.Context, cmd Subscribe) error
}

type subscribeHandler struct{}

func NewAuthorizedSubscribeHandler(logger Logger, metricsClient MetricsClient) SubscribeHandler {
	return subscribeLoggingDecorator{
		base: subscribeMetricsDecorator{
			base: subscribeAuthorizationDecorator{
				base: subscribeHandler{},
			},
			client: metricsClient,
		},
		logger: logger,
	}
}

func NewUnauthorizedSubscribeHandler(logger Logger, metricsClient MetricsClient) SubscribeHandler {
	return subscribeLoggingDecorator{
		base: subscribeMetricsDecorator{
			base:   subscribeHandler{},
			client: metricsClient,
		},
		logger: logger,
	}
}

func NewSubscribeHandler() SubscribeHandler {
	return subscribeHandler{}
}

func (h subscribeHandler) Handle(ctx context.Context, cmd Subscribe) error {
	// Subscribe the user to the newsletter
	return nil
}
