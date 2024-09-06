package main

import (
	"context"
	"database/sql"
	"errors"
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

func (r *PostgresUserRepository) UsePointsForDiscount(ctx context.Context, userID int, points int) error {
	return runInTx(r.db, func(tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx, "SELECT points FROM users WHERE id = $1 FOR UPDATE", userID)

		var currentPoints int
		err := row.Scan(&currentPoints)
		if err != nil {
			return err
		}

		if currentPoints < points {
			return errors.New("not enough points")
		}

		_, err = tx.ExecContext(ctx, "UPDATE users SET points = points - $1 WHERE id = $2", points, userID)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, "UPDATE user_discounts SET next_order_discount = next_order_discount + $1 WHERE user_id = $2", points, userID)
		if err != nil {
			return err
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
