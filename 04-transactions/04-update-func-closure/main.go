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

	usePointsAsDiscountHandler := NewUsePointsAsDiscountHandler(userRepository)

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

type User struct {
	id     int
	email  string
	points int
	cart   *Cart
}

func (u *User) Points() int {
	return u.points
}

func (u *User) Cart() *Cart {
	return u.cart
}

type Cart struct {
	discount int
}

func (c *Cart) Discount() int {
	return c.discount
}

func (u *User) UsePointsAsDiscount(points int) error {
	if points <= 0 {
		return errors.New("points must be greater than 0")
	}

	if points > u.points {
		return errors.New("not enough points")
	}

	u.points -= points
	u.cart.discount += points

	return nil
}

type UserRepository struct {
	db db
}

type db interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (r *UserRepository) Update(ctx context.Context, userID int, updateFn func(user *User) (bool, error)) (err error) {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			err = tx.Commit()
		} else {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				fmt.Println("Rollback failed:", err)
			}
		}
	}()

	row := tx.QueryRowContext(ctx, "SELECT email, points FROM users WHERE id = $1", userID)
	if row.Err() != nil {
		return row.Err()
	}

	var email string
	var currentPoints int
	err = row.Scan(&email, &currentPoints)
	if err != nil {
		return err
	}

	row = tx.QueryRowContext(ctx, "SELECT discount FROM carts WHERE user_id = $1", userID)
	if row.Err() != nil {
		return row.Err()
	}

	var discount int
	err = row.Scan(&discount)
	if err != nil {
		return err
	}

	user := &User{
		id:     userID,
		email:  email,
		points: currentPoints,
		cart: &Cart{
			discount: discount,
		},
	}

	updated, err := updateFn(user)
	if err != nil {
		return err
	}

	if !updated {
		return nil
	}

	_, err = tx.ExecContext(ctx, "UPDATE users SET points = points - $1 WHERE id = $2", user.Points(), userID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "UPDATE carts SET discount = discount + $1 WHERE user_id = $2", user.Cart().Discount(), userID)
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
}

type userRepository interface {
	Update(ctx context.Context, userID int, updateFn func(user *User) (bool, error)) error
}

func NewUsePointsAsDiscountHandler(
	userRepository userRepository,
) UsePointsAsDiscountHandler {
	return UsePointsAsDiscountHandler{
		userRepository: userRepository,
	}
}

func (h UsePointsAsDiscountHandler) Handle(ctx context.Context, cmd UsePointsAsDiscount) error {
	return h.userRepository.Update(ctx, cmd.UserID, func(user *User) (bool, error) {
		err := user.UsePointsAsDiscount(cmd.Points)
		if err != nil {
			return false, err
		}

		return true, nil
	})
}
