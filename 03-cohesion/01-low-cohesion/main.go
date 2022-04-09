package main

import (
	"context"
	"log"
	"os"
)

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
