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
	discountRepository discountRepository
}

type discountRepository interface {
	AddDiscount(ctx context.Context, userID int, discount int) error
}

func NewAddDiscountHandler(
	discountRepository discountRepository,
) AddDiscountHandler {
	return AddDiscountHandler{
		discountRepository: discountRepository,
	}
}

func (h AddDiscountHandler) Handle(ctx context.Context, cmd AddDiscount) error {
	if cmd.Discount <= 0 {
		return errors.New("discount must be greater than 0")
	}

	return h.discountRepository.AddDiscount(ctx, cmd.UserID, cmd.Discount)
}
