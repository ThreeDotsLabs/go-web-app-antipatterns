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
	userRepository userRepository
}

type userRepository interface {
	UpdateByID(ctx context.Context, userID int, updateFn func(user *User) (bool, []any, error)) error
}

func NewUsePointsAsDiscountHandler(
	userRepository userRepository,
) UsePointsAsDiscountHandler {
	return UsePointsAsDiscountHandler{
		userRepository: userRepository,
	}
}

func (h UsePointsAsDiscountHandler) Handle(ctx context.Context, cmd UsePointsAsDiscount) error {
	err := h.userRepository.UpdateByID(ctx, cmd.UserID, func(user *User) (bool, []any, error) {
		err := user.UsePoints(cmd.Points)
		if err != nil {
			return false, nil, err
		}

		event := PointsUsedForDiscount{
			UserID: cmd.UserID,
			Points: cmd.Points,
		}

		return true, []any{event}, nil
	})
	if err != nil {
		return fmt.Errorf("could not update user: %w", err)
	}

	return nil
}
