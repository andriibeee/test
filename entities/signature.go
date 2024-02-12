package entities

import (
	"time"

	"github.com/google/uuid"
)

type Signature struct {
	ID        uuid.UUID
	User      User
	CreatedAt time.Time
}
