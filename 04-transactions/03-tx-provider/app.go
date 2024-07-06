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
	txProvider txProvider
}

type Adapters struct {
	UserRepository userRepository
	CartRepository cartRepository
}

type txProvider interface {
	Transact(txFunc func(adapters Adapters) error) error
}

type userRepository interface {
	GetPoints(ctx context.Context, userID int) (int, error)
	TakePoints(ctx context.Context, userID int, points int) error
}

type cartRepository interface {
	AddDiscount(ctx context.Context, userID int, discount int) error
}

func NewUsePointsAsDiscountHandler(
	txProvider txProvider,
) UsePointsAsDiscountHandler {
	return UsePointsAsDiscountHandler{
		txProvider: txProvider,
	}
}

func (h UsePointsAsDiscountHandler) Handle(ctx context.Context, cmd UsePointsAsDiscount) error {
	return h.txProvider.Transact(func(adapters Adapters) error {
		if cmd.Points <= 0 {
			return errors.New("points must be greater than 0")
		}

		currentPoints, err := adapters.UserRepository.GetPoints(ctx, cmd.UserID)
		if err != nil {
			return fmt.Errorf("could not get points: %w", err)
		}

		if currentPoints < cmd.Points {
			return errors.New("not enough points")
		}

		err = adapters.UserRepository.TakePoints(ctx, cmd.UserID, cmd.Points)
		if err != nil {
			return fmt.Errorf("could not take points: %w", err)
		}

		err = adapters.CartRepository.AddDiscount(ctx, cmd.UserID, cmd.Points)
		if err != nil {
			return fmt.Errorf("could not add discount: %w", err)
		}

		return nil
	})
}
