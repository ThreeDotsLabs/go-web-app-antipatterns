package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"
)

type MetricsClient interface {
	Inc(key string, value int)
}

type Logger interface {
	Println(args ...interface{})
}

func UserFromContext(ctx context.Context) (User, error) {
	u, ok := ctx.Value("user").(User)
	if !ok {
		return User{}, errors.New("could not get user from context")
	}
	return u, nil
}

func ContextWithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, "user", user)
}

type User struct {
	ID     string
	Active bool
}

type Subscribe struct {
	Email        string
	NewsletterID string
}

type SubscribeHandler interface {
	Execute(ctx context.Context, cmd Subscribe) error
}

type subscribeHandler struct{}

func NewSubscribeHandler(logger Logger, metricsClient MetricsClient) SubscribeHandler {
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

func (h subscribeHandler) Execute(ctx context.Context, cmd Subscribe) error {
	// Subscribe the user to the newsletter
	return nil
}

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

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	handler := NewSubscribeHandler(logger, nopMetricsClient{})

	unauthorizedHandler := NewUnauthorizedSubscribeHandler(logger, nopMetricsClient{})

	cmd := Subscribe{
		Email:        "user@example.com",
		NewsletterID: "product-news",
	}

	ctx := ContextWithUser(context.Background(), User{
		ID:     "1000",
		Active: true,
	})

	err := handler.Execute(ctx, cmd)
	if err != nil {
		log.Fatal(err)
	}

	err = unauthorizedHandler.Execute(context.Background(), cmd)
	if err != nil {
		log.Fatal(err)
	}
}

type nopMetricsClient struct{}

func (c nopMetricsClient) Inc(key string, value int) {}
