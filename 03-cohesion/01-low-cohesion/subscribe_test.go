package main

import (
	"context"
	"log"
	"os"
	"testing"
)

func TestSubscribe(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	metricsClient := nopMetricsClient{}

	handler := NewSubscribeHandler(logger, metricsClient)

	user := User{
		ID:     "1000",
		Active: true,
	}

	ctx := ContextWithUser(context.Background(), user)

	cmd := Subscribe{
		Email:        "user@example.com",
		NewsletterID: "product-news",
	}

	err := handler.Handle(ctx, cmd)
	if err != nil {
		t.Fatal(err)
	}
}
