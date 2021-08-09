package internal

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/ThreeDotsLabs/go-web-app-antipatterns/01-coupling/04-loosely-coupled-app-layer/models"
	"github.com/go-sql-driver/mysql"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

type UserStorage struct {
	db *sql.DB
}

func NewUserStorage(db *sql.DB) UserStorage {
	return UserStorage{
		db: db,
	}
}

func (s UserStorage) All(ctx context.Context) ([]User, error) {
	dbUsers, err := models.Users(qm.Load(models.UserRels.Emails)).All(ctx, s.db)
	if err != nil {
		return nil, err
	}

	var users []User
	for _, u := range dbUsers {
		users = append(users, dbUserToApp(u))
	}

	return users, nil
}

func (s UserStorage) ByID(ctx context.Context, id int) (User, error) {
	dbUser, err := models.Users(qm.Load(models.UserRels.Emails), qm.Where("id = ?", id)).One(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}

	return dbUserToApp(dbUser), nil
}

func (s UserStorage) Add(ctx context.Context, user User) (err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			err = tx.Commit()
		} else {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				log.Println("Error while rolling back:", err)
			}
		}
	}()

	dbUser := dbUserFromApp(user)

	err = dbUser.Insert(ctx, tx, boil.Infer())
	if err != nil {
		return err
	}

	dbEmail := dbEmailFromApp(user.Emails()[0])
	dbEmail.UserID = dbUser.ID

	err = dbEmail.Insert(ctx, tx, boil.Infer())
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrEmailAlreadyExists
		}
		return err
	}

	return nil
}

func (s UserStorage) Update(ctx context.Context, user User) error {
	dbUser := dbUserFromApp(user)
	_, err := dbUser.Update(ctx, s.db, boil.Whitelist(models.UserColumns.FirstName, models.UserColumns.LastName))
	return err
}

func (s UserStorage) Delete(ctx context.Context, id int) error {
	_, err := models.Users(qm.Where("id = ?", id)).DeleteAll(ctx, s.db)
	return err
}

func dbUserFromApp(u User) *models.User {
	return &models.User{
		ID:           int64(u.ID()),
		FirstName:    u.FirstName(),
		LastName:     u.LastName(),
		PasswordHash: null.String{},
		LastIP:       null.String{},
		CreatedAt:    null.Time{},
		UpdatedAt:    null.Time{},
	}
}

func dbUserToApp(u *models.User) User {
	var emails []Email
	for _, e := range u.R.Emails {
		emails = append(emails, UnmarshalEmail(e.Address, e.Primary))
	}

	return UnmarshalUser(int(u.ID), u.FirstName, u.LastName, emails)
}

func dbEmailFromApp(e Email) *models.Email {
	return &models.Email{
		Address: e.Address(),
		Primary: e.Primary(),
	}
}
