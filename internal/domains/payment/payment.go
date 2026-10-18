package payment

import "uuid"

type Payment struct {
	ID        uuid.UUID
	OrderID   uuid.UUID
	InvoiceID *string
	Status    Status
	Comment   *string
}
