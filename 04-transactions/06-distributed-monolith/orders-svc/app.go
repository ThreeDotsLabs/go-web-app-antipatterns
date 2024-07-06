package main

import (
	"context"
	"errors"
)

type AddDiscount struct {
	UserID   int
	Discount int
}

type AddDiscountHandler struct {
	cartRepository cartRepository
}

type cartRepository interface {
	AddDiscount(ctx context.Context, userID int, discount int) error
}

func NewAddDiscountHandler(
	cartRepository cartRepository,
) AddDiscountHandler {
	return AddDiscountHandler{
		cartRepository: cartRepository,
	}
}

func (h AddDiscountHandler) Handle(ctx context.Context, cmd AddDiscount) error {
	if cmd.Discount <= 0 {
		return errors.New("discount must be greater than 0")
	}

	return h.cartRepository.AddDiscount(ctx, cmd.UserID, cmd.Discount)
}
