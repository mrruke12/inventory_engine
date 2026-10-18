package product

import "uuid"

type Product struct {
	ID       uuid.UUID
	Name     string
	Quantity int
}
