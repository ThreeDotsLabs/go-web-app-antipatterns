package main

import (
	"context"
	"database/sql"

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

	router, err := NewEventsRouter("redis-b:6379", addDiscountHandler)
	if err != nil {
		panic(err)
	}

	err = router.Run(context.Background())
	if err != nil {
		panic(err)
	}
}
