package internal

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
)

type UserHandler struct {
	storage UserStorage
}

func NewUserHandler(storage UserStorage) UserHandler {
	return UserHandler{
		storage: storage,
	}
}

func (h UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.storage.All(r.Context())
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var usersResponse []UserResponse
	for _, u := range users {
		usersResponse = append(usersResponse, newUserResponse(u))
	}

	err = json.NewEncoder(w).Encode(usersResponse)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h UserHandler) GetUser(w http.ResponseWriter, r *http.Request, rawUserID UserID) {
	userID, err := strconv.Atoi(string(rawUserID))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := h.storage.ByID(r.Context(), userID)
	if err != nil {
		log.Println(err)
		if errors.Is(err, ErrUserNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	userResponse := newUserResponse(user)

	err = json.NewEncoder(w).Encode(userResponse)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h UserHandler) PostUser(w http.ResponseWriter, r *http.Request) {
	var postUserRequest PostUserRequest
	err := json.NewDecoder(r.Body).Decode(&postUserRequest)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user, err := NewUser(postUserRequest.FirstName, postUserRequest.LastName, postUserRequest.Email)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.storage.Add(r.Context(), user)
	if err != nil {
		log.Println(err)
		if errors.Is(err, ErrEmailAlreadyExists) {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h UserHandler) PatchUser(w http.ResponseWriter, r *http.Request, rawUserID UserID) {
	userID, err := strconv.Atoi(string(rawUserID))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var patchUserRequest PatchUserRequest
	err = json.NewDecoder(r.Body).Decode(&patchUserRequest)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	emptyRequest := PatchUserRequest{}
	if patchUserRequest == emptyRequest {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	user, err := h.storage.ByID(r.Context(), userID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = user.ChangeName(patchUserRequest.FirstName, patchUserRequest.LastName)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.storage.Update(r.Context(), user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request, rawUserID UserID) {
	userID, err := strconv.Atoi(string(rawUserID))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.storage.Delete(r.Context(), userID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func newUserResponse(u User) UserResponse {
	var emails []EmailResponse
	for _, e := range u.Emails() {
		emails = append(emails, newEmailResponse(e))
	}

	return UserResponse{
		Id:          u.ID(),
		FirstName:   u.FirstName(),
		LastName:    u.LastName(),
		DisplayName: u.DisplayName(),
		Emails:      emails,
	}
}

func newEmailResponse(e Email) EmailResponse {
	return EmailResponse{
		Address: e.Address(),
		Primary: e.Primary(),
	}
}
