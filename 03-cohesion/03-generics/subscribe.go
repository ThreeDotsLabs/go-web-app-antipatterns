package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

type CommandHandler[C any] interface {
	Execute(ctx context.Context, cmd C) error
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

func (H SubscribeHandler) Execute(ctx context.Context, cmd Subscribe) error {
	// Subscribe the user to the newsletter
	return nil
}

type loggingDecorator[C any] struct {
	base   CommandHandler[C]
	logger Logger
}

func (d loggingDecorator[C]) Execute(ctx context.Context, cmd C) (err error) {
	d.logger.Println("Executing command", commandName(cmd), cmd)
	defer func() {
		if err == nil {
			log.Println("Command executed successfully")
		} else {
			log.Println("Failed to execute command:", err)
		}
	}()

	return d.base.Execute(ctx, cmd)
}

type metricsDecorator[C any] struct {
	base   CommandHandler[C]
	client MetricsClient
}

func (d metricsDecorator[C]) Execute(ctx context.Context, cmd C) (err error) {
	start := time.Now()
	defer func() {
		end := time.Now().Sub(start)
		d.client.Inc(fmt.Sprintf("commands.%s.duration", commandName(cmd)), int(end.Seconds()))

		if err == nil {
			d.client.Inc(fmt.Sprintf("commands.%s.success", commandName(cmd)), 1)
		} else {
			d.client.Inc(fmt.Sprintf("commands.%s.failure", commandName(cmd)), 1)
		}
	}()

	return d.base.Execute(ctx, cmd)
}

type authorizationDecorator[C any] struct {
	base CommandHandler[C]
}

func (d authorizationDecorator[C]) Execute(ctx context.Context, cmd C) error {
	user, err := UserFromContext(ctx)
	if err != nil {
		return err
	}

	if !user.Active {
		return errors.New("the user's account is not active")
	}

	return d.base.Execute(ctx, cmd)
}

func commandName(cmd any) string {
	return strings.ToLower(strings.Split(fmt.Sprintf("%T", cmd), ".")[1])
}
