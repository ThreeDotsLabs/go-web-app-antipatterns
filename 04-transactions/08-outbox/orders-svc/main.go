package main

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "postgres://postgres:postgres@postgres-orders:5432/postgres?sslmode=disable")
	if err != nil {
		panic(err)
	}

	err = MigrateDB(db)
	if err != nil {
		panic(err)
	}

	cartRepo := NewCartRepository(db)

	addDiscountHandler := NewAddDiscountHandler(cartRepo)

	router, err := NewEventsRouter("nats://nats:4222", addDiscountHandler)
	if err != nil {
		panic(err)
	}

	err = router.Run(context.Background())
	if err != nil {
		panic(err)
	}
}
