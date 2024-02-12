package database

import (
	"context"
	"database/sql"

	"github.com/andriibeee/task/entities"
)

type QuestionsService interface {
	Persist(ctx context.Context, questions []entities.Question) error
}

type QuestionsServiceImpl struct {
}

func NewQuestionsService(db *sql.DB) QuestionsService {
	return QuestionsServiceImpl{}
}

func (q QuestionsServiceImpl) Persist(ctx context.Context, questions []entities.Question) error {
	tx := getTransaction(ctx)
	for _, question := range questions {
		_, err := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO questions(uuid, question) VALUES(?, ?);`,
			question.ID.String(),
			question.Question,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
