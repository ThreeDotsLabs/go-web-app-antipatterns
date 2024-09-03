package main

import (
	"context"
	"database/sql"
)

func MigrateDB(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			points INT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS user_discounts (
			user_id INT PRIMARY KEY REFERENCES users(id),
			next_order_discount INT NOT NULL DEFAULT 0
	    );
	`)
	return err
}

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}

func (r *PostgresUserRepository) GetPoints(ctx context.Context, userID int) (int, error) {
	row := r.db.QueryRowContext(ctx, "SELECT points FROM users WHERE id = $1", userID)

	var points int
	err := row.Scan(&points)
	if err != nil {
		return 0, err
	}

	return points, nil
}

func (r *PostgresUserRepository) TakePoints(ctx context.Context, userID int, points int) error {
	_, err := r.db.ExecContext(ctx, "UPDATE users SET points = points - $1 WHERE id = $2", points, userID)
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
	return err
}
