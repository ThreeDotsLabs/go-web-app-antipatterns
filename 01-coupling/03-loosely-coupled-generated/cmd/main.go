package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/ThreeDotsLabs/go-web-app-antipatterns/01-coupling/03-loosely-coupled-generated/internal"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root@tcp(mysql)/loosely_coupled_app_layer?parseTime=true")
	if err != nil {
		log.Fatal("failed to connect to the database")
	}

	storage := internal.NewUserStorage(db)
	h := internal.NewUserHandler(storage)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	handler := internal.HandlerFromMux(h, r)

	err = http.ListenAndServe(":8080", handler)
	if err != nil {
		log.Fatal(err)
	}
}
