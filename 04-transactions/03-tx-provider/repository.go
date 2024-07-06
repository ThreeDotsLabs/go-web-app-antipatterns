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

		CREATE TABLE IF NOT EXISTS carts (
			user_id INT PRIMARY KEY REFERENCES users(id),
			discount INT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	    );
	`)
	return err
}

type db interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type UserRepository struct {
	db db
}

func NewUserRepository(db db) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) GetPoints(ctx context.Context, userID int) (int, error) {
	row := r.db.QueryRowContext(ctx, "SELECT points FROM users WHERE id = $1 FOR UPDATE", userID)
	if row.Err() != nil {
		return 0, row.Err()
	}

	var points int
	err := row.Scan(&points)
	if err != nil {
		return 0, err
	}

	return points, nil
}

func (r *UserRepository) TakePoints(ctx context.Context, userID int, points int) error {
	_, err := r.db.ExecContext(ctx, "UPDATE users SET points = points - $1 WHERE id = $2", points, userID)
	return err
}

type CartRepository struct {
	db db
}

func NewCartRepository(db db) *CartRepository {
	return &CartRepository{
		db: db,
	}
}

func (r *CartRepository) AddDiscount(ctx context.Context, userID int, discount int) error {
	_, err := r.db.ExecContext(ctx, "UPDATE carts SET discount = discount + $1 WHERE user_id = $2", discount, userID)
	return err
}
