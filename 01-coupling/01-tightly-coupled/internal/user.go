package internal

import (
	"time"

	"github.com/go-playground/validator/v10"
)

type User struct {
	ID           int        `json:"id" gorm:"primaryKey"`
	FirstName    string     `json:"first_name" validate:"required_without=LastName"`
	LastName     string     `json:"last_name" validate:"required_without=FirstName"`
	DisplayName  string     `json:"display_name"`
	Email        string     `json:"email,omitempty" gorm:"-"`
	Emails       []Email    `json:"emails" validate:"required,dive" gorm:"constraint:OnDelete:CASCADE"`
	PasswordHash string     `json:"-"`
	LastIP       string     `json:"-"`
	CreatedAt    *time.Time `json:"-"`
	UpdatedAt    *time.Time `json:"-"`
}

func (u *User) Validate() error {
	validate := validator.New()
	return validate.Struct(u)
}

func (u *User) SetDisplayName() {
	if u.FirstName != "" {
		u.DisplayName = u.FirstName

		if u.LastName != "" {
			u.DisplayName += " " + u.LastName
		}
	} else if u.LastName != "" {
		u.DisplayName = u.LastName
	}
}

type Email struct {
	ID      int    `json:"-" gorm:"primaryKey"`
	Address string `json:"address" validate:"required,email" gorm:"size:256;uniqueIndex"`
	Primary bool   `json:"primary"`
	UserID  int    `json:"-"`
}
