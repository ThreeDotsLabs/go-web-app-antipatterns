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

	userRepo := NewUserRepository(db)
	cartRepo := NewCartRepository(db)

	usePointsAsDiscountHandler := NewUsePointsAsDiscountHandler(db, userRepo, cartRepo)

	handler := NewHTTPHandler(usePointsAsDiscountHandler)

	err = http.ListenAndServe(":8080", handler)
	if err != nil {
		panic(err)
	}
}
