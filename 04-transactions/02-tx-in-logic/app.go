package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type UsePointsAsDiscount struct {
	UserID int
	Points int
}

type UsePointsAsDiscountHandler struct {
	db             *sql.DB
	userRepository userRepository
	cartRepository cartRepository
}

type userRepository interface {
	GetPoints(ctx context.Context, tx *sql.Tx, userID int) (int, error)
	TakePoints(ctx context.Context, tx *sql.Tx, userID int, points int) error
}

type cartRepository interface {
	AddDiscount(ctx context.Context, tx *sql.Tx, userID int, discount int) error
}

func NewUsePointsAsDiscountHandler(
	db *sql.DB,
	userRepository userRepository,
	cartRepository cartRepository,
) UsePointsAsDiscountHandler {
	return UsePointsAsDiscountHandler{
		db:             db,
		userRepository: userRepository,
		cartRepository: cartRepository,
	}
}

func (h UsePointsAsDiscountHandler) Handle(ctx context.Context, cmd UsePointsAsDiscount) error {
	return runInTx(h.db, func(tx *sql.Tx) error {
		if cmd.Points <= 0 {
			return errors.New("points must be greater than 0")
		}

		currentPoints, err := h.userRepository.GetPoints(ctx, tx, cmd.UserID)
		if err != nil {
			return fmt.Errorf("could not get points: %w", err)
		}

		if currentPoints < cmd.Points {
			return errors.New("not enough points")
		}

		err = h.userRepository.TakePoints(ctx, tx, cmd.UserID, cmd.Points)
		if err != nil {
			return fmt.Errorf("could not take points: %w", err)
		}

		err = h.cartRepository.AddDiscount(ctx, tx, cmd.UserID, cmd.Points)
		if err != nil {
			return fmt.Errorf("could not add discount: %w", err)
		}

		return nil
	})
}

func runInTx(db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	err = fn(tx)
	if err == nil {
		return tx.Commit()
	}

	rollbackErr := tx.Rollback()
	if rollbackErr != nil {
		return errors.Join(err, rollbackErr)
	}

	return err
}
