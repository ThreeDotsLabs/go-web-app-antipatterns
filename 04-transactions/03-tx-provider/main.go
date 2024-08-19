package main

import (
	"database/sql"
	"errors"
	"net/http"

	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "postgres://postgres:postgres@postgres:5432/postgres?sslmode=disable")
	if err != nil {
		panic(err)
	}

	err = MigrateDB(db)
	if err != nil {
		panic(err)
	}

	txProvider := NewTransactionProvider(db)

	usePointsAsDiscountHandler := NewUsePointsAsDiscountHandler(txProvider)

	handler := NewHTTPHandler(usePointsAsDiscountHandler)

	err = http.ListenAndServe(":8080", handler)
	if err != nil {
		panic(err)
	}
}

type TransactionProvider struct {
	db *sql.DB
}

func NewTransactionProvider(db *sql.DB) *TransactionProvider {
	return &TransactionProvider{
		db: db,
	}
}

func (p *TransactionProvider) Transact(txFunc func(adapters Adapters) error) error {
	return runInTx(p.db, func(tx *sql.Tx) error {
		adapters := Adapters{
			UserRepository: NewUserRepository(tx),
			DiscountRepository: NewDiscountRepository(tx),
		}

		return txFunc(adapters)
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
