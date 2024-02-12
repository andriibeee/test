package commands

import (
	"context"
	"errors"

	"github.com/andriibeee/task/database"
	"github.com/andriibeee/task/entities"
	"github.com/google/uuid"
)

var AccessForbidden = errors.New("access forbidden")

type FetchInput struct {
	User      entities.User
	Signature uuid.UUID
}

type FetchOutput struct {
	Answers   []entities.Answer `json:"answers"`
	Timestamp int64             `json:"timestamp"`
}

// ...

type FetchCommand interface {
	Handle(ctx context.Context, input FetchInput) (*FetchOutput, error)
}

type FetchCommandImpl struct {
	answers    database.AnswersService
	signatures database.SignaturesService
}

func NewFetchCommand(
	answers database.AnswersService,
	signatures database.SignaturesService,
) FetchCommand {
	return FetchCommandImpl{
		answers:    answers,
		signatures: signatures,
	}
}

func (cmd FetchCommandImpl) Handle(ctx context.Context, input FetchInput) (*FetchOutput, error) {
	signature, err := cmd.signatures.Get(ctx, input.Signature)
	if err != nil {
		return nil, err
	}

	if signature.User.ID != input.User.ID {
		return nil, AccessForbidden
	}

	answers, err := cmd.answers.Get(ctx, signature.ID)
	if err != nil {
		return nil, err
	}

	return &FetchOutput{
		Answers:   answers,
		Timestamp: signature.CreatedAt.Unix(),
	}, nil
}
