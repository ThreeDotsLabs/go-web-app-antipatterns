package internal

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/ThreeDotsLabs/go-web-app-antipatterns/01-coupling/03-loosely-coupled-generated/models"
	"github.com/go-playground/validator/v10"
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
		usersResponse = append(usersResponse, userResponseFromDBModel(u))
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

	userResponse := userResponseFromDBModel(user)

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

	type createRequest struct {
		Email     string `validate:"required,email"`
		FirstName string `validate:"required_without=LastName"`
		LastName  string `validate:"required_without=FirstName"`
	}

	validate := validator.New()
	err = validate.Struct(createRequest(postUserRequest))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user := &models.User{
		FirstName: postUserRequest.FirstName,
		LastName:  postUserRequest.LastName,
	}
	email := &models.Email{Address: postUserRequest.Email}

	err = h.storage.Add(r.Context(), user, email)
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

	type updateRequest struct {
		FirstName *string `validate:"required_without=LastName"`
		LastName  *string `validate:"required_without=FirstName"`
	}

	validate := validator.New()
	err = validate.Struct(updateRequest(patchUserRequest))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := h.storage.ByID(r.Context(), userID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if patchUserRequest.FirstName != nil {
		user.FirstName = *patchUserRequest.FirstName
	}

	if patchUserRequest.LastName != nil {
		user.LastName = *patchUserRequest.LastName
	}

	if displayName(user.FirstName, user.LastName) == "" {
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

func userResponseFromDBModel(u *models.User) UserResponse {
	var emails []EmailResponse
	for _, e := range u.R.Emails {
		emails = append(emails, emailResponseFromDBModel(e))
	}

	return UserResponse{
		Id:          int(u.ID),
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		DisplayName: displayName(u.FirstName, u.LastName),
		Emails:      emails,
	}
}

func emailResponseFromDBModel(e *models.Email) EmailResponse {
	return EmailResponse{
		Address: e.Address,
		Primary: e.Primary,
	}
}

func displayName(firstName string, lastName string) string {
	if firstName != "" {
		name := firstName

		if lastName != "" {
			name += " " + lastName
		}

		return name
	} else if lastName != "" {
		return lastName
	}

	return ""
}
