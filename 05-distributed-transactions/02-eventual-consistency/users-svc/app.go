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
	PublishPointsUsedForDiscount(ctx context.Context, userID int, points int) error
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

	err = h.eventPublisher.PublishPointsUsedForDiscount(ctx, cmd.UserID, cmd.Points)
	if err != nil {
		return fmt.Errorf("could not publish event: %w", err)
	}

	return nil
}
