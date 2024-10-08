package main

import (
	"context"
	"database/sql"
)

func MigrateDB(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS user_discounts (
			user_id INT PRIMARY KEY,
			next_order_discount INT NOT NULL DEFAULT 0
	    );
	`)
	return err
}

type PostgresDiscountRepository struct {
	db *sql.DB
}

func NewPostgresDiscountRepository(db *sql.DB) *PostgresDiscountRepository {
	return &PostgresDiscountRepository{
		db: db,
	}
}

func (r *PostgresDiscountRepository) AddDiscount(ctx context.Context, userID int, discount int) error {
	_, err := r.db.ExecContext(ctx, "UPDATE user_discounts SET next_order_discount = next_order_discount + $1 WHERE user_id = $2", discount, userID)
	if err != nil {
		return err
	}

	return nil
}
