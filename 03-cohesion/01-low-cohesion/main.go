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

func (h SubscribeHandler) Execute(ctx context.Context, cmd Subscribe) (err error) {
	start := time.Now()
	h.logger.Println("Subscribing to newsletter", cmd)
	defer func() {
		end := time.Now().Sub(start)
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

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	handler := NewSubscribeHandler(logger, nopMetricsClient{})

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
}

type nopMetricsClient struct{}

func (c nopMetricsClient) Inc(key string, value int) {}
