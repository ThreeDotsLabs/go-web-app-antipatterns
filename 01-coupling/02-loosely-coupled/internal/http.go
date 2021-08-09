package internal

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"

	"github.com/go-chi/chi/v5"
)

type CreateUserRequest struct {
	FirstName string `json:"first_name" validate:"required_without=LastName"`
	LastName  string `json:"last_name" validate:"required_without=FirstName"`
	Email     string `json:"email" validate:"required,email"`
}

type UpdateUserRequest struct {
	FirstName *string `json:"first_name" validate:"required_without=LastName"`
	LastName  *string `json:"last_name" validate:"required_without=FirstName"`
}

type UserResponse struct {
	ID          int             `json:"id"`
	FirstName   string          `json:"first_name"`
	LastName    string          `json:"last_name"`
	DisplayName string          `json:"display_name"`
	Emails      []EmailResponse `json:"emails"`
}

type EmailResponse struct {
	Address string `json:"address"`
	Primary bool   `json:"primary"`
}

type UserHandler struct {
	storage UserStorage
}

func NewUserHandler(storage UserStorage) UserHandler {
	return UserHandler{
		storage: storage,
	}
}

func (h UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.storage.All()
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

func (h UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := h.storage.ByID(userID)
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
	var createUserRequest CreateUserRequest
	err := json.NewDecoder(r.Body).Decode(&createUserRequest)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	validate := validator.New()
	err = validate.Struct(createUserRequest)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user := userDBModelFromCreateRequest(createUserRequest)

	err = h.storage.Add(user)
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

func (h UserHandler) PatchUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var updateUserRequest UpdateUserRequest
	err = json.NewDecoder(r.Body).Decode(&updateUserRequest)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	emptyRequest := UpdateUserRequest{}
	if updateUserRequest == emptyRequest {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	validate := validator.New()
	err = validate.Struct(updateUserRequest)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := h.storage.ByID(userID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if updateUserRequest.FirstName != nil {
		user.FirstName = *updateUserRequest.FirstName
	}

	if updateUserRequest.LastName != nil {
		user.LastName = *updateUserRequest.LastName
	}

	if displayName(user.FirstName, user.LastName) == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.storage.Update(user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.storage.Delete(userID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func userResponseFromDBModel(u UserDBModel) UserResponse {
	var emails []EmailResponse
	for _, e := range u.Emails {
		emails = append(emails, emailResponseFromDBModel(e))
	}

	return UserResponse{
		ID:          u.ID,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		DisplayName: displayName(u.FirstName, u.LastName),
		Emails:      emails,
	}
}

func emailResponseFromDBModel(e EmailDBModel) EmailResponse {
	return EmailResponse{
		Address: e.Address,
		Primary: e.Primary,
	}
}

func userDBModelFromCreateRequest(r CreateUserRequest) UserDBModel {
	return UserDBModel{
		FirstName: r.FirstName,
		LastName:  r.LastName,
		Emails: []EmailDBModel{
			{
				Address: r.Email,
			},
		},
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
