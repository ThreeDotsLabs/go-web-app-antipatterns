package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	Name           string
	URL            string
	DiscountDBPort int
	UsersDBPort    int
}

func TestAll(t *testing.T) {
	testCases := []testCase{
		{Name: "01-no-tx", URL: "http://localhost:8101"},
		{Name: "02-tx-in-logic", URL: "http://localhost:8102"},
		{Name: "03-tx-provider", URL: "http://localhost:8103"},
		{Name: "04-tx-in-repo", URL: "http://localhost:8104"},
		{Name: "05-update-func-closure", URL: "http://localhost:8105"},
		{Name: "06-distributed-monolith", URL: "http://localhost:8162", DiscountDBPort: 5433, UsersDBPort: 5434},
		{Name: "07-eventual-consistency", URL: "http://localhost:8172", DiscountDBPort: 5433, UsersDBPort: 5434},
		{Name: "08-outbox", URL: "http://localhost:8182", DiscountDBPort: 5433, UsersDBPort: 5434},
	}
	for _, a := range testCases {
		t.Run(a.Name, func(t *testing.T) {
			t.Parallel()

			userID := createUser(t, a, 100)

			usePoints(t, a, userID, 25)

			assertPoints(t, a, userID, 75)
			assertDiscount(t, a, userID, 25)
		})
	}
}

var dbs = map[int]*sql.DB{}

var lock sync.Mutex

func getDB(t *testing.T, port int) *sql.DB {
	t.Helper()

	if port == 0 {
		port = 5432
	}

	lock.Lock()
	defer lock.Unlock()

	if dbs[port] != nil {
		return dbs[port]
	}

	db, err := sql.Open("postgres", fmt.Sprintf("postgres://postgres:postgres@localhost:%v/postgres?sslmode=disable", port))
	require.NoError(t, err)

	dbs[port] = db

	return db
}

func createUser(t *testing.T, tc testCase, points int) int {
	t.Helper()

	email := uuid.NewString() + "@" + tc.Name + ".com"

	usersDB := getDB(t, tc.UsersDBPort)

	row := usersDB.QueryRow("INSERT INTO users (email, points) VALUES ($1, $2) RETURNING id", email, points)

	var id int
	err := row.Scan(&id)
	require.NoError(t, err)

	discountDB := getDB(t, tc.DiscountDBPort)

	_, err = discountDB.Exec("INSERT INTO discounts (user_id) VALUES ($1)", id)
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

func assertPoints(t *testing.T, tc testCase, userID int, expectedPoints int) {
	t.Helper()

	usersDB := getDB(t, tc.UsersDBPort)

	assert.EventuallyWithT(t, func(t *assert.CollectT) {
		row := usersDB.QueryRowContext(context.Background(), "SELECT points FROM users WHERE id = $1", userID)

		var points int
		err := row.Scan(&points)
		require.NoError(t, err)

		assert.Equal(t, expectedPoints, points)
	}, 2*time.Second, 100*time.Millisecond)
}

func assertDiscount(t *testing.T, tc testCase, userID int, expectedDiscount int) {
	t.Helper()

	discountDB := getDB(t, tc.DiscountDBPort)

	assert.EventuallyWithT(t, func(t *assert.CollectT) {
		row := discountDB.QueryRowContext(context.Background(), "SELECT next_order_discount FROM discounts WHERE user_id = $1", userID)

		var discount int
		err := row.Scan(&discount)
		require.NoError(t, err)

		assert.Equal(t, expectedDiscount, discount)
	}, 2*time.Second, 100*time.Millisecond)
}
