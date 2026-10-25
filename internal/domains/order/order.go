package order

import (
	"time"
)

type Order struct {
	ID        string
	Status    Status
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewOrder(orderID string) *Order {
	return &Order{
		ID:     orderID,
		Status: StatusPendingPayment,
	}
}
