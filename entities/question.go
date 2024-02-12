package entities

import "github.com/google/uuid"

type Question struct {
	ID       uuid.UUID `json:"id"`
	Question string    `json:"question"`
}
