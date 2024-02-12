package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/andriibeee/task/entities"
	"github.com/google/uuid"
)

var SignatureNotFound = errors.New("signature not found")

type SignaturesService interface {
	Persist(ctx context.Context, sign entities.Signature) error
	Get(ctx context.Context, id uuid.UUID) (*entities.Signature, error)
}

type SignaturesServiceImpl struct {
	db *sql.DB
}

func NewSignaturesService(db *sql.DB) SignaturesService {
	return SignaturesServiceImpl{db: db}
}

func (service SignaturesServiceImpl) Persist(ctx context.Context, sign entities.Signature) error {
	tx := getTransaction(ctx)
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO signatures
		(uuid, user_id, timestamp)
		VALUES (?, ?, ?);`,
		sign.ID.String(),
		sign.User.ID.String(),
		sign.CreatedAt.Unix(),
	)
	return err
}

func (service SignaturesServiceImpl) Get(ctx context.Context, id uuid.UUID) (*entities.Signature, error) {
	row := service.db.QueryRowContext(
		ctx,
		`SELECT signatures.uuid, 
				signatures.user_id,
				users.user_name,
				signatures.timestamp
		 FROM signatures
		 INNER JOIN users ON signatures.user_id = users.uuid
		 WHERE signatures.uuid = ?`,
		id.String(),
	)

	var signID string
	var userID string
	var userName string
	var timestamp int64

	err := row.Err()
	if err != nil {
		return nil, err
	}

	err = row.Scan(
		&signID,
		&userID,
		&userName,
		&timestamp,
	)

	if err != nil {
		return nil, err
	}

	return &entities.Signature{
		ID: uuid.MustParse(signID),
		User: entities.User{
			ID:       uuid.MustParse(userID),
			UserName: userName,
		},
		CreatedAt: time.Unix(timestamp, 0),
	}, nil
}
