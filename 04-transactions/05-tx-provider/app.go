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
	txProvider txProvider
}

type Adapters struct {
	UserRepository     userRepository
	AuditLogRepository auditLogRepository
}

type txProvider interface {
	Transact(txFunc func(adapters Adapters) error) error
}

type userRepository interface {
	UpdateByID(ctx context.Context, userID int, updateFn func(user *User) (bool, error)) error
}

type auditLogRepository interface {
	StoreAuditLog(ctx context.Context, log string) error
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
		err := adapters.UserRepository.UpdateByID(ctx, cmd.UserID, func(user *User) (bool, error) {
			err := user.UsePointsAsDiscount(cmd.Points)
			if err != nil {
				return false, err
			}

			return true, nil
		})
		if err != nil {
			return fmt.Errorf("could not use points as discount: %w", err)
		}

		log := fmt.Sprintf("used %d points as discount for user %d", cmd.Points, cmd.UserID)
		err = adapters.AuditLogRepository.StoreAuditLog(ctx, log)
		if err != nil {
			return fmt.Errorf("could not store audit log: %w", err)
		}

		return nil
	})
}
