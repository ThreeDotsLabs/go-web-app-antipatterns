package main

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type UsePointsAsDiscount struct {
	UserID int
	Points int
}

type UsePointsAsDiscountHandler struct {
	userRepository UserRepository
}

type UserRepository interface {
	UpdateByID(ctx context.Context, userID int, updateFn func(user *User) (bool, []cqrs.Event, error)) error
}

func NewUsePointsAsDiscountHandler(
	userRepository UserRepository,
) UsePointsAsDiscountHandler {
	return UsePointsAsDiscountHandler{
		userRepository: userRepository,
	}
}

func (h UsePointsAsDiscountHandler) Handle(ctx context.Context, cmd UsePointsAsDiscount) error {
	return h.userRepository.UpdateByID(ctx, cmd.UserID, func(user *User) (bool, []cqrs.Event, error) {
		err := user.UsePoints(cmd.Points)
		if err != nil {
			return false, nil, err
		}

		event := PointsUsedForDiscount{
			UserID: cmd.UserID,
			Points: cmd.Points,
		}

		return true, []cqrs.Event{event}, nil
	})
}
