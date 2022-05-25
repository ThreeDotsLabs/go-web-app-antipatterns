package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type loggingDecorator[C any] struct {
	base   CommandHandler[C]
	logger Logger
}

func (d loggingDecorator[C]) Handle(ctx context.Context, cmd C) (err error) {
	d.logger.Println("Executing command", commandName(cmd), cmd)
	defer func() {
		if err == nil {
			d.logger.Println("Command executed successfully")
		} else {
			d.logger.Println("Failed to execute command:", err)
		}
	}()

	return d.base.Handle(ctx, cmd)
}

type metricsDecorator[C any] struct {
	base   CommandHandler[C]
	client MetricsClient
}

func (d metricsDecorator[C]) Handle(ctx context.Context, cmd C) (err error) {
	start := time.Now()
	defer func() {
		end := time.Since(start)
		d.client.Inc(fmt.Sprintf("commands.%s.duration", commandName(cmd)), int(end.Seconds()))

		if err == nil {
			d.client.Inc(fmt.Sprintf("commands.%s.success", commandName(cmd)), 1)
		} else {
			d.client.Inc(fmt.Sprintf("commands.%s.failure", commandName(cmd)), 1)
		}
	}()

	return d.base.Handle(ctx, cmd)
}

type authorizationDecorator[C any] struct {
	base CommandHandler[C]
}

func (d authorizationDecorator[C]) Handle(ctx context.Context, cmd C) error {
	user, err := UserFromContext(ctx)
	if err != nil {
		return err
	}

	if !user.Active {
		return errors.New("the user's account is not active")
	}

	return d.base.Handle(ctx, cmd)
}

func commandName(cmd any) string {
	return strings.ToLower(strings.Split(fmt.Sprintf("%T", cmd), ".")[1])
}
