package main

import (
	"context"
	"fmt"
)

type UsePointsAsDiscount struct {
	UserID int
	Points int
}

type UsePointsAsDiscountHandler struct {
	userRepository UserRepository
	ordersService  OrdersService
}

type UserRepository interface {
	UpdateByID(ctx context.Context, userID int, updateFn func(user *User) (bool, error)) error
}

type OrdersService interface {
	AddDiscount(ctx context.Context, userID int, discount int) error
}

func NewUsePointsAsDiscountHandler(
	userRepository UserRepository,
	ordersService OrdersService,
) UsePointsAsDiscountHandler {
	return UsePointsAsDiscountHandler{
		userRepository: userRepository,
		ordersService:  ordersService,
	}
}

func (h UsePointsAsDiscountHandler) Handle(ctx context.Context, cmd UsePointsAsDiscount) error {
	err := h.userRepository.UpdateByID(ctx, cmd.UserID, func(user *User) (bool, error) {
		err := user.UsePoints(cmd.Points)
		if err != nil {
			return false, err
		}

		return true, nil
	})
	if err != nil {
		return fmt.Errorf("could not update user: %w", err)
	}

	err = h.ordersService.AddDiscount(ctx, cmd.UserID, cmd.Points)
	if err != nil {
		return fmt.Errorf("could not add discount: %w", err)
	}

	return nil
}
