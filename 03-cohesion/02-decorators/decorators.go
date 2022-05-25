package main

import (
	"context"
	"errors"
	"time"
)

type subscribeLoggingDecorator struct {
	base   SubscribeHandler
	logger Logger
}

func (d subscribeLoggingDecorator) Handle(ctx context.Context, cmd Subscribe) (err error) {
	d.logger.Println("Subscribing to newsletter", cmd)
	defer func() {
		if err == nil {
			d.logger.Println("Subscribed to newsletter")
		} else {
			d.logger.Println("Failed subscribing to newsletter:", err)
		}
	}()

	return d.base.Handle(ctx, cmd)
}

type subscribeMetricsDecorator struct {
	base   SubscribeHandler
	client MetricsClient
}

func (d subscribeMetricsDecorator) Handle(ctx context.Context, cmd Subscribe) (err error) {
	start := time.Now()
	defer func() {
		end := time.Since(start)
		d.client.Inc("commands.subscribe.duration", int(end.Seconds()))

		if err == nil {
			d.client.Inc("commands.subscribe.success", 1)
		} else {
			d.client.Inc("commands.subscribe.failure", 1)
		}
	}()

	return d.base.Handle(ctx, cmd)
}

type subscribeAuthorizationDecorator struct {
	base SubscribeHandler
}

func (d subscribeAuthorizationDecorator) Handle(ctx context.Context, cmd Subscribe) error {
	user, err := UserFromContext(ctx)
	if err != nil {
		return err
	}

	if !user.Active {
		return errors.New("the user's account is not active")
	}

	return d.base.Handle(ctx, cmd)
}
