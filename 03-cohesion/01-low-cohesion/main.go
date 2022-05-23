package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	handler := NewSubscribeHandler(logger, nopMetricsClient{})

	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		var request SubscribeHTTPRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		cmd := Subscribe{
			Email:        request.Email,
			NewsletterID: request.NewsletterID,
		}

		user, err := userFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ctx := ContextWithUser(r.Context(), user)

		err = handler.Execute(ctx, cmd)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	eventHandler := func(ctx context.Context, event UserSignedUp) error {
		if !event.ProductNewsConsent {
			return nil
		}

		fakeUser := User{
			ID:     event.ID,
			Active: true,
		}

		ctx = ContextWithUser(ctx, fakeUser)

		cmd := Subscribe{
			Email:        event.Email,
			NewsletterID: "product-news",
		}

		return handler.Execute(ctx, cmd)
	}

	rpcHandler := func(ctx context.Context, req SubscribeRPCRequest) error {
		fakeUser := User{
			ID:     "1", // Missing ID in the context, let's assume it's the root user making changes
			Active: true,
		}

		ctx = ContextWithUser(ctx, fakeUser)

		cmd := Subscribe{
			Email:        req.Email,
			NewsletterID: "product-news",
		}

		return handler.Execute(ctx, cmd)
	}

	_ = httpHandler
	_ = eventHandler
	_ = rpcHandler
}

func userFromRequest(r *http.Request) (User, error) {
	token := r.Header.Get("Authorization")

	// Verify the token, create the user struct out of it
	_ = token

	user := User{
		ID:     "1000",
		Active: true,
	}

	return user, nil
}

type SubscribeHTTPRequest struct {
	Email        string `json:"email"`
	NewsletterID string `json:"newsletter_id"`
}

type SubscribeRPCRequest struct {
	Email        string
	NewsletterID string
}

type UserSignedUp struct {
	ID                 string
	Email              string
	ProductNewsConsent bool
}

type nopMetricsClient struct{}

func (c nopMetricsClient) Inc(key string, value int) {}
