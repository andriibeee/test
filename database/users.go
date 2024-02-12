package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/andriibeee/task/entities"
	"github.com/google/uuid"
)

var UserNotFound = errors.New("user not found")

type UsersService interface {
	Persist(ctx context.Context, user entities.User) error
	Get(ctx context.Context, id uuid.UUID) (*entities.User, error)
}

type UsersServiceImpl struct {
	db *sql.DB
}

func NewUsersService(db *sql.DB) UsersService {
	return UsersServiceImpl{db: db}
}

func (service UsersServiceImpl) Get(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	row := service.db.QueryRowContext(ctx, `SELECT user_name FROM users WHERE uuid = ?;`, id.String())

	user := entities.User{
		ID: id,
	}

	err := row.Scan(&user.UserName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, UserNotFound
	}
	return &user, err
}

func (service UsersServiceImpl) Persist(ctx context.Context, user entities.User) error {
	_, err := service.db.ExecContext(
		ctx,
		`INSERT INTO users(uuid, user_name)
		VALUES(?,?);`,
		user.ID.String(),
		user.UserName,
	)

	return err
}
