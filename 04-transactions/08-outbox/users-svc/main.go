package main

import (
	"context"
	"database/sql"
	"net/http"

	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "postgres://postgres:postgres@postgres-users:5432/postgres?sslmode=disable")
	if err != nil {
		panic(err)
	}

	err = MigrateDB(db)
	if err != nil {
		panic(err)
	}

	userRepo := NewUserRepository(db)

	usePointsAsDiscountHandler := NewUsePointsAsDiscountHandler(userRepo)

	forwarder, err := NewEventsForwarder("redis-b:6379", db)
	if err != nil {
		panic(err)
	}

	go func() {
		err := forwarder.Run(context.Background())
		if err != nil {
			panic(err)
		}
	}()

	handler := NewHTTPHandler(usePointsAsDiscountHandler)

	err = http.ListenAndServe(":8080", handler)
	if err != nil {
		panic(err)
	}
}
