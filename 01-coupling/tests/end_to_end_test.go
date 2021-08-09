package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser(t *testing.T) {
	services := []struct {
		Name string
		Port int
	}{
		{
			Name: "01_tightly_coupled",
			Port: 8080,
		},
		{
			Name: "02_loosely_coupled",
			Port: 8081,
		},
		{
			Name: "03_loosely_coupled_generated",
			Port: 8082,
		},
		{
			Name: "04_loosely_coupled_app_layer",
			Port: 8083,
		},
	}

	testCases := []struct {
		Name     string
		TestFunc func(*testing.T, HTTPClient)
	}{
		{
			Name:     "user_lifecycle",
			TestFunc: testUserLifecycle,
		},
		{
			Name:     "create_user",
			TestFunc: testCreateUser,
		},
		{
			Name:     "update_user",
			TestFunc: testUpdateUser,
		},
	}

	for i := range services {
		s := services[i]
		for j := range testCases {
			tc := testCases[j]

			name := fmt.Sprintf("%v_%v", s.Name, tc.Name)
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				client := NewHTTPClient(t, s.Port)
				tc.TestFunc(t, client)
			})
		}
	}
}

func testUserLifecycle(t *testing.T, client HTTPClient) {
	firstName := gofakeit.FirstName()
	lastName := gofakeit.LastName()
	email := gofakeit.Email()

	_, ok := findUserByEmail(client.GetAllUsers(), email)
	require.False(t, ok, "Didn't expect to find the user by email")

	client.PostUser(firstName, lastName, email, http.StatusCreated)

	user, ok := findUserByEmail(client.GetAllUsers(), email)
	require.True(t, ok, "Expected to find the user by email")

	require.Equal(t, firstName+" "+lastName, user.DisplayName)
	require.Len(t, user.Emails, 1)

	userByID, ok := client.GetUser(user.ID)
	require.True(t, ok, "Expected to find the user by ID")

	require.Equal(t, firstName+" "+lastName, user.DisplayName)
	require.Len(t, userByID.Emails, 1)

	client.DeleteUser(user.ID)

	_, ok = findUserByEmail(client.GetAllUsers(), email)
	require.False(t, ok, "Didn't expect to find the user by email")

	_, ok = client.GetUser(user.ID)
	require.False(t, ok, "Didn't expect to find the user by ID")

	client.PostUser(firstName, lastName, email, http.StatusCreated)
}

func testCreateUser(t *testing.T, client HTTPClient) {
	firstName := gofakeit.FirstName()
	lastName := gofakeit.LastName()

	testCases := []struct {
		Name                string
		FirstName           string
		LastName            string
		Email               string
		ExpectedDisplayName string
		ExpectedStatusCode  int
	}{
		{
			Name:               "empty_request",
			FirstName:          "",
			LastName:           "",
			Email:              "",
			ExpectedStatusCode: http.StatusBadRequest,
		},
		{
			Name:               "empty_email",
			FirstName:          firstName,
			LastName:           lastName,
			Email:              "",
			ExpectedStatusCode: http.StatusBadRequest,
		},
		{
			Name:               "empty_first_and_last_name",
			FirstName:          "",
			LastName:           "",
			Email:              gofakeit.Email(),
			ExpectedStatusCode: http.StatusBadRequest,
		},
		{
			Name:               "empty_last_name",
			FirstName:          firstName,
			LastName:           "",
			Email:              gofakeit.Email(),
			ExpectedStatusCode: http.StatusCreated,
		},
		{
			Name:               "empty_first_name",
			FirstName:          "",
			LastName:           lastName,
			Email:              gofakeit.Email(),
			ExpectedStatusCode: http.StatusCreated,
		},
		{
			Name:                "all_provided",
			FirstName:           firstName,
			LastName:            lastName,
			Email:               gofakeit.Email(),
			ExpectedDisplayName: firstName + " " + lastName,
			ExpectedStatusCode:  http.StatusCreated,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			client.PostUser(tc.FirstName, tc.LastName, tc.Email, tc.ExpectedStatusCode)

			if tc.ExpectedDisplayName != "" {
				user, ok := findUserByEmail(client.GetAllUsers(), tc.Email)
				require.True(t, ok, "Expected to find the user by email")
				assert.Equal(t, tc.ExpectedDisplayName, user.DisplayName)
			}
		})
	}

	// Add the same email the second time
	email := gofakeit.Email()
	client.PostUser(firstName, lastName, email, http.StatusCreated)
	client.PostUser(firstName, lastName, email, http.StatusBadRequest)
}

func testUpdateUser(t *testing.T, client HTTPClient) {
	firstName := gofakeit.FirstName()
	lastName := gofakeit.LastName()
	email := gofakeit.Email()

	client.PostUser(firstName, lastName, email, http.StatusCreated)

	user, ok := findUserByEmail(client.GetAllUsers(), email)
	require.True(t, ok, "Expected to find the user by email")

	emptyName := ""
	newFirstName := gofakeit.FirstName()
	newLastName := gofakeit.LastName()

	testCases := []struct {
		Name                string
		NewFirstName        *string
		NewLastName         *string
		ExpectedStatusCode  int
		ExpectedDisplayName string
	}{
		{
			Name:                "no_changes",
			NewFirstName:        nil,
			NewLastName:         nil,
			ExpectedStatusCode:  http.StatusNoContent,
			ExpectedDisplayName: firstName + " " + lastName,
		},
		{
			Name:                "change_first_name",
			NewFirstName:        &newFirstName,
			NewLastName:         nil,
			ExpectedStatusCode:  http.StatusNoContent,
			ExpectedDisplayName: newFirstName + " " + lastName,
		},
		{
			Name:                "change_last_name",
			NewFirstName:        nil,
			NewLastName:         &newLastName,
			ExpectedStatusCode:  http.StatusNoContent,
			ExpectedDisplayName: newFirstName + " " + newLastName,
		},
		{
			Name:                "change_first_and_last_name",
			NewFirstName:        &firstName,
			NewLastName:         &lastName,
			ExpectedStatusCode:  http.StatusNoContent,
			ExpectedDisplayName: firstName + " " + lastName,
		},
		{
			Name:                "delete_last_name",
			NewFirstName:        nil,
			NewLastName:         &emptyName,
			ExpectedStatusCode:  http.StatusNoContent,
			ExpectedDisplayName: firstName,
		},
		{
			Name:                "delete_first_name_when_no_last_name",
			NewFirstName:        &emptyName,
			NewLastName:         nil,
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedDisplayName: firstName,
		},
		{
			Name:                "delete_first_name_and_add_last_name",
			NewFirstName:        &emptyName,
			NewLastName:         &newLastName,
			ExpectedStatusCode:  http.StatusNoContent,
			ExpectedDisplayName: newLastName,
		},
		{
			Name:                "delete_last_name_and_add_first_name",
			NewFirstName:        &newFirstName,
			NewLastName:         &emptyName,
			ExpectedStatusCode:  http.StatusNoContent,
			ExpectedDisplayName: newFirstName,
		},
		{
			Name:                "delete_first_and_last_name",
			NewFirstName:        &emptyName,
			NewLastName:         &emptyName,
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedDisplayName: newFirstName,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			// No parallelism allowed here, as the order matters
			client.PatchUser(user.ID, tc.NewFirstName, tc.NewLastName, tc.ExpectedStatusCode)
			updatedUser, ok := client.GetUser(user.ID)
			require.True(t, ok, "Expected to find the user by ID")

			assert.Equal(t, tc.ExpectedDisplayName, updatedUser.DisplayName)
		})
	}
}

func findUserByEmail(users []User, emailToFind string) (User, bool) {
	for _, u := range users {
		for _, e := range u.Emails {
			if e.Address == emailToFind {
				return u, true
			}
		}
	}

	return User{}, false
}
