package main

import (
	"database/sql"
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

	discountRepo := NewPostgresDiscountRepository(db)

	addDiscountHandler := NewAddDiscountHandler(discountRepo)

	handler := NewHTTPHandler(addDiscountHandler)

	err = http.ListenAndServe(":8080", handler)
	if err != nil {
		panic(err)
	}
}
