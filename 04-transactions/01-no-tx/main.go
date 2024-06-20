package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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

	userRepository := NewUserRepository(db)
	cartRepository := NewCartRepository(db)

	usePointsAsDiscountHandler := NewUsePointsAsDiscountHandler(userRepository, cartRepository)

	http.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
		type payload struct {
			Email string `json:"email"`
		}

		var p payload
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = userRepository.Create(r.Context(), p.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})

	http.HandleFunc("POST /use-points", func(w http.ResponseWriter, r *http.Request) {
		type payload struct {
			UserID int `json:"user_id"`
			Points int `json:"points"`
		}

		var p payload
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		cmd := UsePointsAsDiscount{
			UserID: p.UserID,
			Points: p.Points,
		}

		err = usePointsAsDiscountHandler.Handle(r.Context(), cmd)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	})

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

type UserRepository struct {
	db *sql.DB
}

func (r *UserRepository) GetPoints(ctx context.Context, userID int) (int, error) {
	row := r.db.QueryRowContext(ctx, "SELECT points FROM users WHERE id = $1", userID)
	if row.Err() != nil {
		return 0, row.Err()
	}

	var points int
	err := row.Scan(&points)
	if err != nil {
		return 0, err
	}

	return points, nil
}

func (r *UserRepository) TakePoints(ctx context.Context, userID int, points int) error {
	_, err := r.db.ExecContext(ctx, "UPDATE users SET points = points - $1 WHERE id = $2", points, userID)
	return err
}

func (r *UserRepository) Create(ctx context.Context, email string) error {
	res, err := r.db.ExecContext(ctx, "INSERT INTO users (email) VALUES ($1)", email)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, "INSERT INTO carts (user_id) VALUES ($1)", id)
	return err
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

type CartRepository struct {
	db *sql.DB
}

func (r *CartRepository) AddDiscount(ctx context.Context, userID int, discount int) error {
	_, err := r.db.ExecContext(ctx, "UPDATE carts SET discount = discount + $1 WHERE user_id = $2", discount, userID)
	return err
}

func NewCartRepository(db *sql.DB) *CartRepository {
	return &CartRepository{
		db: db,
	}
}

func MigrateDB(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			points INT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS carts (
			user_id INT PRIMARY KEY REFERENCES users(id),
			discount INT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	    );
	`)
	return err
}

type UsePointsAsDiscount struct {
	UserID int
	Points int
}

type UsePointsAsDiscountHandler struct {
	userRepository userRepository
	cartRepository cartRepository
}

type userRepository interface {
	GetPoints(ctx context.Context, userID int) (int, error)
	TakePoints(ctx context.Context, userID int, points int) error
}

type cartRepository interface {
	AddDiscount(ctx context.Context, userID int, discount int) error
}

func NewUsePointsAsDiscountHandler(
	userRepository userRepository,
	cartRepository cartRepository,
) UsePointsAsDiscountHandler {
	return UsePointsAsDiscountHandler{
		userRepository: userRepository,
		cartRepository: cartRepository,
	}
}

func (h UsePointsAsDiscountHandler) Handle(ctx context.Context, cmd UsePointsAsDiscount) error {
	if cmd.Points <= 0 {
		return errors.New("points must be greater than 0")
	}

	points, err := h.userRepository.GetPoints(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("could not get points: %w", err)
	}

	if points > cmd.Points {
		return errors.New("not enough points")
	}

	err = h.userRepository.TakePoints(ctx, cmd.UserID, cmd.Points)
	if err != nil {
		return fmt.Errorf("could not take points: %w", err)
	}

	err = h.cartRepository.AddDiscount(ctx, cmd.UserID, points)
	if err != nil {
		return fmt.Errorf("could not add discount: %w", err)
	}

	return nil
}
