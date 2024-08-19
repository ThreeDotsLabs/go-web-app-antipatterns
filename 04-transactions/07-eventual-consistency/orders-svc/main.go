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

	discountRepo := NewDiscountRepository(db)

	addDiscountHandler := NewAddDiscountHandler(discountRepo)

	router, err := NewEventsRouter("redis-a:6379", addDiscountHandler)
	if err != nil {
		panic(err)
	}

	err = router.Run(context.Background())
	if err != nil {
		panic(err)
	}
}
