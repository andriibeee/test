package entities

import "github.com/google/uuid"

type Answer struct {
	ID       uuid.UUID `json:"id"`
	Question Question  `json:"question"`
	Answer   string    `json:"answer"`
}
