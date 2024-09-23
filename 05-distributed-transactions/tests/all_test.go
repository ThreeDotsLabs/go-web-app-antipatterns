package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	Name string
	URL  string
}

func TestAll(t *testing.T) {
	testCases := []testCase{
		{Name: "01-distributed-monolith", URL: "http://localhost:8101"},
		{Name: "02-eventual-consistency", URL: "http://localhost:8103"},
		{Name: "03-outbox", URL: "http://localhost:8105"},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			userID := createUser(t, tc, 100)

			usePoints(t, tc, userID, 25)

			assertPoints(t, userID, 75)
			assertDiscount(t, userID, 25)

			usePoints(t, tc, userID, 50)

			assertPoints(t, userID, 25)
			assertDiscount(t, userID, 75)
		})
	}
}

func getDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("postgres", fmt.Sprintf("postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"))
	require.NoError(t, err)

	return db
}

func createUser(t *testing.T, tc testCase, points int) int {
	t.Helper()

	email := uuid.NewString() + "@" + tc.Name + ".com"

	usersDB := getDB(t)

	row := usersDB.QueryRow("INSERT INTO users (email, points) VALUES ($1, $2) RETURNING id", email, points)

	var id int
	err := row.Scan(&id)
	require.NoError(t, err)

	discountDB := getDB(t)

	_, err = discountDB.Exec("INSERT INTO user_discounts (user_id) VALUES ($1)", id)
	require.NoError(t, err)

	return int(id)
}

func usePoints(t *testing.T, tc testCase, userID int, points int) {
	t.Helper()

	type request struct {
		UserID int `json:"user_id"`
		Points int `json:"points"`
	}

	r := request{
		UserID: userID,
		Points: points,
	}

	payload, err := json.Marshal(r)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", tc.URL+"/use-points", bytes.NewReader(payload))
	require.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, res.StatusCode, string(body))
}

func assertPoints(t *testing.T, userID int, expectedPoints int) {
	t.Helper()

	usersDB := getDB(t)

	assert.EventuallyWithT(t, func(t *assert.CollectT) {
		row := usersDB.QueryRowContext(context.Background(), "SELECT points FROM users WHERE id = $1", userID)

		var points int
		err := row.Scan(&points)
		require.NoError(t, err)

		assert.Equal(t, expectedPoints, points)
	}, 2*time.Second, 100*time.Millisecond)
}

func assertDiscount(t *testing.T, userID int, expectedDiscount int) {
	t.Helper()

	discountDB := getDB(t)

	assert.EventuallyWithT(t, func(t *assert.CollectT) {
		row := discountDB.QueryRowContext(context.Background(), "SELECT next_order_discount FROM user_discounts WHERE user_id = $1", userID)

		var discount int
		err := row.Scan(&discount)
		require.NoError(t, err)

		assert.Equal(t, expectedDiscount, discount)
	}, 2*time.Second, 100*time.Millisecond)
}
