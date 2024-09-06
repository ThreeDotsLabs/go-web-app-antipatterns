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

		CREATE TABLE IF NOT EXISTS audit_log (
			id SERIAL PRIMARY KEY,
			log TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return err
}

type db interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type PostgresUserRepository struct {
	db db
}

func NewPostgresUserRepository(db db) *PostgresUserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}
func (r *PostgresUserRepository) UpdateByID(ctx context.Context, userID int, updateFn func(user *User) (bool, error)) error {
	row := r.db.QueryRowContext(ctx, "SELECT email, points FROM users WHERE id = $1 FOR UPDATE", userID)

	var email string
	var currentPoints int
	err := row.Scan(&email, &currentPoints)
	if err != nil {
		return err
	}

	row = r.db.QueryRowContext(ctx, "SELECT next_order_discount FROM user_discounts WHERE user_id = $1 FOR UPDATE", userID)

	var discount int
	err = row.Scan(&discount)
	if err != nil {
		return err
	}

	discounts := UnmarshalDiscounts(discount)
	user := UnmarshalUser(userID, email, currentPoints, discounts)

	updated, err := updateFn(user)
	if err != nil {
		return err
	}

	if !updated {
		return nil
	}

	_, err = r.db.ExecContext(ctx, "UPDATE users SET email = $1, points = $2 WHERE id = $3", user.Email(), user.Points(), user.ID())
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, "UPDATE user_discounts SET next_order_discount = $1 WHERE user_id = $2", user.Discounts().NextOrderDiscount(), user.ID())
	if err != nil {
		return err
	}

	return nil
}

type PostgresAuditLogRepository struct {
	db db
}

func NewPostgresAuditLogRepository(db db) *PostgresAuditLogRepository {
	return &PostgresAuditLogRepository{
		db: db,
	}
}

func (r *PostgresAuditLogRepository) StoreAuditLog(ctx context.Context, log string) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO audit_log (log) VALUES ($1)", log)
	return err
}
