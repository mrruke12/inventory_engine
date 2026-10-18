package order

import "uuid"

type Item struct {
	ID        uuid.UUID
	ProductID uuid.UUID
	Quantity  int
}
