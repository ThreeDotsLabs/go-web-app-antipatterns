package internal

import (
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

type UserDBModel struct {
	ID           int            `gorm:"column:id;primaryKey"`
	FirstName    string         `gorm:"column:first_name"`
	LastName     string         `gorm:"column:last_name"`
	Emails       []EmailDBModel `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	PasswordHash string         `gorm:"column:password_hash"`
	LastIP       string         `gorm:"column:last_ip"`
	CreatedAt    *time.Time     `gorm:"column:created_at"`
	UpdatedAt    *time.Time     `gorm:"column:updated_at"`
}

func (UserDBModel) TableName() string {
	return "users"
}

type EmailDBModel struct {
	ID      int    `gorm:"column:id;primaryKey"`
	Address string `gorm:"column:address;size:256;uniqueIndex"`
	Primary bool   `gorm:"column:primary"`
	UserID  int    `gorm:"column:user_id"`
}

func (EmailDBModel) TableName() string {
	return "emails"
}

type UserStorage struct {
	db *gorm.DB
}

func NewUserStorage(db *gorm.DB) UserStorage {
	return UserStorage{
		db: db,
	}
}

func (s UserStorage) All() ([]UserDBModel, error) {
	var users []UserDBModel

	result := s.db.Preload("Emails").Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	return users, nil
}

func (s UserStorage) ByID(id int) (UserDBModel, error) {
	var user UserDBModel

	result := s.db.Preload("Emails").First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return UserDBModel{}, ErrUserNotFound
		}
		return UserDBModel{}, result.Error
	}

	return user, nil
}

func (s UserStorage) Add(user UserDBModel) error {
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

func (s UserStorage) Update(user UserDBModel) error {
	result := s.db.Save(user)
	return result.Error
}

func (s UserStorage) Delete(id int) error {
	result := s.db.Delete(&UserDBModel{}, id)
	return result.Error
}
