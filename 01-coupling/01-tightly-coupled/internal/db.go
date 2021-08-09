package internal

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

type UserStorage struct {
	db *gorm.DB
}

func NewUserStorage(db *gorm.DB) UserStorage {
	return UserStorage{
		db: db,
	}
}

func (s UserStorage) All() ([]User, error) {
	var users []User

	result := s.db.Preload("Emails").Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	return users, nil
}

func (s UserStorage) ByID(id int) (User, error) {
	var user User

	result := s.db.Preload("Emails").First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return User{}, ErrUserNotFound
		}
		return User{}, result.Error
	}

	return user, nil
}

func (s UserStorage) Add(user User) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Omit("Emails").Create(&user)
		if result.Error != nil {
			return result.Error
		}

		email := &user.Emails[0]
		email.UserID = user.ID

		result = tx.Create(&email)
		if result.Error != nil {
			var mysqlErr *mysql.MySQLError
			if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 {
				return ErrEmailAlreadyExists
			}
			return result.Error
		}
		return nil
	})
}

func (s UserStorage) Update(user User) error {
	result := s.db.Save(user)
	return result.Error
}

func (s UserStorage) Delete(id int) error {
	result := s.db.Delete(&User{}, id)
	return result.Error
}
