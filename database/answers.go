package database

import (
	"context"
	"database/sql"

	"github.com/andriibeee/task/entities"
	"github.com/google/uuid"
)

type AnswersService interface {
	Persist(ctx context.Context, id uuid.UUID, answers []entities.Answer) error
	Get(ctx context.Context, signature uuid.UUID) ([]entities.Answer, error)
}

type AnswersServiceImpl struct {
	db *sql.DB
}

func NewAnswersService(db *sql.DB) AnswersService {
	return AnswersServiceImpl{db: db}
}

func (answ AnswersServiceImpl) Persist(ctx context.Context, id uuid.UUID, answers []entities.Answer) error {
	tx := getTransaction(ctx)
	for _, answer := range answers {
		answerID := uuid.New()
		_, err := tx.ExecContext(ctx,
			`INSERT INTO answers(uuid, question_id, answer, signature_id)
			  VALUES(?, ?, ?, ?)`,
			answerID.String(),
			answer.Question.ID.String(),
			answer.Answer,
			id.String(),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (answ AnswersServiceImpl) Get(ctx context.Context, signature uuid.UUID) ([]entities.Answer, error) {
	rows, err := answ.db.QueryContext(ctx,
		`SELECT answers.uuid, answers.question_id,
		questions.question, answers.answer
		FROM answers
		INNER JOIN questions ON answers.question_id = questions.uuid
		WHERE signature_id = ?;`, signature.String())
	if err != nil {
		return nil, err
	}

	var answers []entities.Answer

	for rows.Next() {

		var id string
		var questionUUID string
		var question string
		var answer string

		err := rows.Scan(
			&id,
			&questionUUID,
			&question,
			&answer,
		)

		if err != nil {
			return nil, err
		}

		answers = append(answers, entities.Answer{
			ID: uuid.MustParse(id),
			Question: entities.Question{
				ID:       uuid.MustParse(questionUUID),
				Question: question,
			},
			Answer: answer,
		})
	}

	return answers, nil

}
