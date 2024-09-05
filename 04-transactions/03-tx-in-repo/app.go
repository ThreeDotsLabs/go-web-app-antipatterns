package main

import (
	"context"
	"errors"
	"fmt"
)

type UsePointsAsDiscount struct {
	UserID int
	Points int
}

type UsePointsAsDiscountHandler struct {
	userRepository UserRepository
}

type UserRepository interface {
	UsePointsForDiscount(ctx context.Context, userID int, point int) error
}

func NewUsePointsAsDiscountHandler(
	userRepository UserRepository,
) UsePointsAsDiscountHandler {
	return UsePointsAsDiscountHandler{
		userRepository: userRepository,
	}
}

func (h UsePointsAsDiscountHandler) Handle(ctx context.Context, cmd UsePointsAsDiscount) error {
	if cmd.Points <= 0 {
		return errors.New("points must be greater than 0")
	}

	err := h.userRepository.UsePointsForDiscount(ctx, cmd.UserID, cmd.Points)
	if err != nil {
		return fmt.Errorf("could not use points as discount: %w", err)
	}

	return nil
}
