package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

type HTTPClient struct {
	t       *testing.T
	client  http.Client
	baseURL *url.URL
}

func NewHTTPClient(t *testing.T, port int) HTTPClient {
	client := http.Client{}

	baseUrl, err := url.Parse(fmt.Sprintf("http://localhost:%v", port))
	require.NoError(t, err)

	return HTTPClient{
		t:       t,
		client:  client,
		baseURL: baseUrl,
	}
}

type User struct {
	ID          int     `json:"id"`
	DisplayName string  `json:"display_name"`
	Emails      []Email `json:"emails"`
}

type PostUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type PatchUserRequest struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
}

type Email struct {
	Address string `json:"address"`
	Primary bool   `json:"primary"`
}

func (c HTTPClient) GetAllUsers() []User {
	resp, err := c.client.Get(c.relativeURL("/users"))
	require.NoError(c.t, err)

	require.Equal(c.t, http.StatusOK, resp.StatusCode)

	defer func() {
		_ = resp.Body.Close()
	}()

	var users []User
	err = json.NewDecoder(resp.Body).Decode(&users)
	require.NoError(c.t, err)

	return users
}

func (c HTTPClient) GetUser(id int) (User, bool) {
	resp, err := c.client.Get(c.relativeURL(fmt.Sprintf("/users/%v", id)))
	require.NoError(c.t, err)

	if resp.StatusCode == http.StatusNotFound {
		return User{}, false
	}

	require.Equal(c.t, http.StatusOK, resp.StatusCode)

	defer func() {
		_ = resp.Body.Close()
	}()

	var user User
	err = json.NewDecoder(resp.Body).Decode(&user)
	require.NoError(c.t, err)

	return user, true
}

func (c HTTPClient) PostUser(firstName string, lastName string, email string, expectedStatusCode int) {
	postUserRequest := PostUserRequest{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}

	body := &bytes.Buffer{}

	err := json.NewEncoder(body).Encode(postUserRequest)
	require.NoError(c.t, err)

	resp, err := c.client.Post(c.relativeURL(fmt.Sprintf("/users")), "application/json", body)
	require.NoError(c.t, err)

	require.Equal(c.t, expectedStatusCode, resp.StatusCode)
}

func (c HTTPClient) PatchUser(id int, firstName *string, lastName *string, expectedStatusCode int) {
	patchuserRequest := PatchUserRequest{
		FirstName: firstName,
		LastName:  lastName,
	}
	body := &bytes.Buffer{}

	err := json.NewEncoder(body).Encode(patchuserRequest)
	require.NoError(c.t, err)

	req, err := http.NewRequest("PATCH", c.relativeURL(fmt.Sprintf("/users/%v", id)), body)
	require.NoError(c.t, err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	require.NoError(c.t, err)

	require.Equal(c.t, expectedStatusCode, resp.StatusCode)
}

func (c HTTPClient) DeleteUser(id int) {
	req, err := http.NewRequest("DELETE", c.relativeURL(fmt.Sprintf("/users/%v", id)), nil)
	require.NoError(c.t, err)

	resp, err := c.client.Do(req)
	require.NoError(c.t, err)

	require.Equal(c.t, http.StatusNoContent, resp.StatusCode)
}

func (c HTTPClient) relativeURL(path string) string {
	return c.baseURL.ResolveReference(&url.URL{Path: path}).String()
}
