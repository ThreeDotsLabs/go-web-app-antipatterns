package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

type OrdersClient struct {
	client *http.Client
	url    string
}

func NewOrdersClient(url string) *OrdersClient {
	return &OrdersClient{
		client: &http.Client{},
		url:    url,
	}
}

type addDiscountBody struct {
	UserID   int `json:"user_id"`
	Discount int `json:"discount"`
}

func (c OrdersClient) AddDiscount(ctx context.Context, userID int, discount int) error {
	fullURL := c.url + "/add-discount"

	payload := addDiscountBody{
		UserID:   userID,
		Discount: discount,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return err
	}

	return nil
}
