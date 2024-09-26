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
	eventPublisher EventPublisher
}

type UserRepository interface {
	UpdateByID(ctx context.Context, userID int, updateFn func(user *User) (bool, error)) error
}

type EventPublisher interface {
	Publish(ctx context.Context, event any) error
}

type PointsUsedForDiscount struct {
	UserID int `json:"user_id"`
	Points int `json:"points"`
}

func NewUsePointsAsDiscountHandler(
	userRepository UserRepository,
	eventPublisher EventPublisher,
) UsePointsAsDiscountHandler {
	return UsePointsAsDiscountHandler{
		userRepository: userRepository,
		eventPublisher: eventPublisher,
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

	event := PointsUsedForDiscount{
		UserID: cmd.UserID,
		Points: cmd.Points,
	}

	err = h.eventPublisher.Publish(ctx, event)
	if err != nil {
		return fmt.Errorf("could not publish event: %w", err)
	}

	return nil
}
