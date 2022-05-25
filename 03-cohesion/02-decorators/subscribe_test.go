package main

import (
	"context"
	"testing"
)

func TestSubscribe(t *testing.T) {
	handler := NewSubscribeHandler()

	cmd := Subscribe{
		Email:        "user@example.com",
		NewsletterID: "product-news",
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatal(err)
	}
}
