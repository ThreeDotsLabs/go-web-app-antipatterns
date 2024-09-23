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

func (r *PostgresUserRepository) UpdateByID(ctx context.Context, userID int, updateFn func(user *User) (bool, []Event, error)) error {
	return runInTx(r.db, func(tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx, "SELECT email, points FROM users WHERE id = $1 FOR UPDATE", userID)

		var email string
		var currentPoints int
		err := row.Scan(&email, &currentPoints)
		if err != nil {
			return err
		}

		user := UnmarshalUser(userID, email, currentPoints)

		updated, events, err := updateFn(user)
		if err != nil {
			return err
		}

		if !updated {
			return nil
		}

		_, err = tx.ExecContext(ctx, "UPDATE users SET email = $1, points = $2 WHERE id = $3", user.Email(), user.Points(), user.ID())
		if err != nil {
			return err
		}

		publisher, err := NewEventPublisher(tx)
		if err != nil {
			return err
		}

		for _, event := range events {
			err = publisher.Publish(ctx, event)
			if err != nil {
				return err
			}
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
