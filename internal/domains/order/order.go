package order

import (
	"time"
	"uuid"
)

type Order struct {
	ID        uuid.UUID
	Status    Status
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
