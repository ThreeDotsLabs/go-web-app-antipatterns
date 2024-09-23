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
	userRepository     UserRepository
	discountRepository DiscountRepository
}

type UserRepository interface {
	GetPoints(ctx context.Context, userID int) (int, error)
	TakePoints(ctx context.Context, userID int, points int) error
}

type DiscountRepository interface {
	AddDiscount(ctx context.Context, userID int, discount int) error
}

func NewUsePointsAsDiscountHandler(
	userRepository UserRepository,
	discountRepository DiscountRepository,
) UsePointsAsDiscountHandler {
	return UsePointsAsDiscountHandler{
		userRepository:     userRepository,
		discountRepository: discountRepository,
	}
}

func (h UsePointsAsDiscountHandler) Handle(ctx context.Context, cmd UsePointsAsDiscount) error {
	if cmd.Points <= 0 {
		return errors.New("points must be greater than 0")
	}

	currentPoints, err := h.userRepository.GetPoints(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("could not get points: %w", err)
	}

	if currentPoints < cmd.Points {
		return errors.New("not enough points")
	}

	err = h.userRepository.TakePoints(ctx, cmd.UserID, cmd.Points)
	if err != nil {
		return fmt.Errorf("could not take points: %w", err)
	}

	err = h.discountRepository.AddDiscount(ctx, cmd.UserID, cmd.Points)
	if err != nil {
		return fmt.Errorf("could not add discount: %w", err)
	}

	return nil
}
