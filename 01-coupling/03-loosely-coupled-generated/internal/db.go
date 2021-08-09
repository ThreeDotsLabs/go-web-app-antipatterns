package internal

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/ThreeDotsLabs/go-web-app-antipatterns/01-coupling/03-loosely-coupled-generated/models"
	"github.com/go-sql-driver/mysql"
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

func (s UserStorage) All(ctx context.Context) ([]*models.User, error) {
	return models.Users(qm.Load(models.UserRels.Emails)).All(ctx, s.db)
}

func (s UserStorage) ByID(ctx context.Context, id int) (*models.User, error) {
	user, err := models.Users(qm.Load(models.UserRels.Emails), qm.Where("id = ?", id)).One(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (s UserStorage) Add(ctx context.Context, user *models.User, email *models.Email) (err error) {
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

	err = user.Insert(ctx, tx, boil.Infer())
	if err != nil {
		return err
	}

	email.UserID = user.ID

	err = email.Insert(ctx, tx, boil.Infer())
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrEmailAlreadyExists
		}
		return err
	}

	return nil
}

func (s UserStorage) Update(ctx context.Context, user *models.User) error {
	_, err := user.Update(ctx, s.db, boil.Whitelist(models.UserColumns.FirstName, models.UserColumns.LastName))
	return err
}

func (s UserStorage) Delete(ctx context.Context, id int) error {
	_, err := models.Users(qm.Where("id = ?", id)).DeleteAll(ctx, s.db)
	return err
}
