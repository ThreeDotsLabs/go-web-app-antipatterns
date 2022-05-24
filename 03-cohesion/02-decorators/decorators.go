package main

import (
	"context"
	"errors"
	"log"
	"time"
)

type subscribeLoggingDecorator struct {
	base   SubscribeHandler
	logger Logger
}

func (d subscribeLoggingDecorator) Execute(ctx context.Context, cmd Subscribe) (err error) {
	d.logger.Println("Subscribing to newsletter", cmd)
	defer func() {
		if err == nil {
			log.Println("Subscribed to newsletter")
		} else {
			log.Println("Failed subscribing to newsletter:", err)
		}
	}()

	return d.base.Execute(ctx, cmd)
}

type subscribeMetricsDecorator struct {
	base   SubscribeHandler
	client MetricsClient
}

func (d subscribeMetricsDecorator) Execute(ctx context.Context, cmd Subscribe) (err error) {
	start := time.Now()
	defer func() {
		end := time.Now().Sub(start)
		d.client.Inc("commands.subscribe.duration", int(end.Seconds()))

		if err == nil {
			d.client.Inc("commands.subscribe.success", 1)
		} else {
			d.client.Inc("commands.subscribe.failure", 1)
		}
	}()

	return d.base.Execute(ctx, cmd)
}

type subscribeAuthorizationDecorator struct {
	base SubscribeHandler
}

func (d subscribeAuthorizationDecorator) Execute(ctx context.Context, cmd Subscribe) error {
	user, err := UserFromContext(ctx)
	if err != nil {
		return err
	}

	if !user.Active {
		return errors.New("the user's account is not active")
	}

	return d.base.Execute(ctx, cmd)
}