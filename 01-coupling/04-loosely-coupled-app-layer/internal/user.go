package internal

import (
	"errors"
	"strings"
)

var (
	ErrNameRequired  = errors.New("either first name or last name is required")
	ErrEmailRequired = errors.New("email address is required")
	ErrInvalidEmail  = errors.New("invalid email address")
)

type User struct {
	id        int
	firstName string
	lastName  string
	emails    []Email
}

func NewUser(firstName string, lastName string, emailAddress string) (User, error) {
	if firstName == "" && lastName == "" {
		return User{}, ErrNameRequired
	}

	email, err := NewEmail(emailAddress, true)
	if err != nil {
		return User{}, err
	}

	return User{
		firstName: firstName,
		lastName:  lastName,
		emails:    []Email{email},
	}, nil
}

// UnmarshalUser loads the user from database data. It shouldn't be used for anything else.
func UnmarshalUser(id int, firstName string, lastName string, emails []Email) User {
	return User{
		id:        id,
		firstName: firstName,
		lastName:  lastName,
		emails:    emails,
	}
}

func (u User) ID() int {
	return u.id
}

func (u User) FirstName() string {
	return u.firstName
}

func (u User) LastName() string {
	return u.lastName
}

func (u User) Emails() []Email {
	return u.emails
}

func (u User) PrimaryEmail() Email {
	for _, e := range u.emails {
		if e.primary {
			return e
		}
	}

	// Normally, it's not a good practice to panic in the logic code.
	// However, since we're sure properly created User must have a primary e-mail, this is a situation that shouldn't ever happen.
	// In a rare occasion that it happens, we choose to panic. The Recoverer middleware will catch it for us.
	panic("no primary email found")
}

func (u *User) ChangeName(newFirstName *string, newLastName *string) error {
	if newFirstName == nil && newLastName == nil {
		return nil
	}

	firstName := u.firstName
	lastName := u.lastName

	if newFirstName != nil {
		firstName = *newFirstName
	}

	if newLastName != nil {
		lastName = *newLastName
	}

	if firstName == "" && lastName == "" {
		return ErrNameRequired
	}

	u.firstName = firstName
	u.lastName = lastName

	return nil
}

func (u User) DisplayName() string {
	if u.firstName != "" {
		if u.lastName != "" {
			return u.firstName + " " + u.lastName
		}

		return u.firstName
	}

	return u.lastName
}

type Email struct {
	address string
	primary bool
}

func NewEmail(address string, primary bool) (Email, error) {
	if address == "" {
		return Email{}, ErrEmailRequired
	}

	// A naive validation to make the example short, but you get the idea
	if !strings.Contains(address, "@") {
		return Email{}, ErrInvalidEmail
	}

	return Email{
		address: address,
		primary: primary,
	}, nil
}

// UnmarshalEmail loads the e-mail from database data. It shouldn't be used for anything else.
func UnmarshalEmail(address string, primary bool) Email {
	return Email{
		address: address,
		primary: primary,
	}
}

func (e Email) Address() string {
	return e.address
}

func (e Email) Primary() bool {
	return e.primary
}
