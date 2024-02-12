package commands

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/andriibeee/task/database"
	"github.com/andriibeee/task/entities"

	"github.com/google/uuid"
)

type SignInput struct {
	User      entities.User
	Questions map[uuid.UUID]entities.Question
	Answers   map[uuid.UUID]entities.Answer
}

type SignOutput struct {
	Signature string `json:"signature"`
}

// ...

type SignCommand interface {
	Handle(ctx context.Context, input SignInput) (*SignOutput, error)
}

type SignCommandImpl struct {
	users      database.UsersService
	questions  database.QuestionsService
	answers    database.AnswersService
	signatures database.SignaturesService
	tm         database.TransactionManager
}

func NewSignCommand(
	users database.UsersService,
	questions database.QuestionsService,
	answers database.AnswersService,
	signatures database.SignaturesService,
	tm database.TransactionManager,
) SignCommandImpl {
	return SignCommandImpl{
		users:      users,
		questions:  questions,
		answers:    answers,
		signatures: signatures,
		tm:         tm,
	}
}

func (sign SignCommandImpl) EnsureUserExists(ctx context.Context, input SignInput) error {
	_, err := sign.users.Get(ctx, input.User.ID)
	if err != nil && errors.Is(err, database.UserNotFound) {
		return sign.users.Persist(ctx, input.User)
	}
	return err
}

func (sign SignCommandImpl) SanityCheck(ctx context.Context, input SignInput) error {
	for _, question := range input.Questions {
		_, ok := input.Answers[question.ID]
		if !ok {
			return fmt.Errorf("answer for question with id '%s' is missing", question.ID)
		}
	}

	for id, _ := range input.Answers {
		_, ok := input.Questions[id]
		if !ok {
			return fmt.Errorf("question with id '%s' is missing", id)
		}
	}

	return nil
}

func (sign SignCommandImpl) PopulateQuestions(ctx context.Context, input SignInput) error {
	questions := make([]entities.Question, len(input.Questions))

	for _, question := range input.Questions {
		questions = append(questions, question)
	}

	return sign.questions.Persist(ctx, questions)
}

func (sign SignCommandImpl) PopulateAnswers(ctx context.Context, id uuid.UUID, input SignInput) error {
	var answers []entities.Answer

	for _, answer := range input.Answers {
		answers = append(answers, answer)
	}

	return sign.answers.Persist(ctx, id, answers)
}

func (sign SignCommandImpl) CreateSignature(ctx context.Context, input SignInput) (uuid.UUID, error) {
	id := uuid.New()
	now := time.Now()
	err := sign.signatures.Persist(ctx, entities.Signature{
		ID:        id,
		User:      input.User,
		CreatedAt: now,
	})
	return id, err
}

func (sign SignCommandImpl) Handle(ctx context.Context, input SignInput) (*SignOutput, error) {

	err := sign.EnsureUserExists(ctx, input)
	if err != nil {
		return nil, err
	}

	err = sign.SanityCheck(ctx, input)
	if err != nil {
		return nil, err
	}

	ctx, err = sign.tm.Begin(ctx)
	if err != nil {
		return nil, err
	}

	err = sign.PopulateQuestions(ctx, input)
	if err != nil {
		sign.tm.Rollback(ctx)
		return nil, err
	}

	id, err := sign.CreateSignature(ctx, input)
	if err != nil {
		sign.tm.Rollback(ctx)
		return nil, err
	}

	err = sign.PopulateAnswers(ctx, id, input)
	if err != nil {
		sign.tm.Rollback(ctx)
		return nil, err
	}

	sign.tm.Commit(ctx)

	return &SignOutput{
		Signature: id.String(),
	}, nil
}
