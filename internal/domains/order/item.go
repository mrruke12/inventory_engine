package order

import (
	"uuid"

	"github.com/mrruke12/inventory_engine/pkg/aggr"
)

type Item struct {
	ID        uuid.UUID
	OrderID   uuid.UUID
	ProductID uuid.UUID
	Quantity  int
}

func NewItem(orderID, productID uuid.UUID, quantity int) *Item {
	return &Item{
		ID:        uuid.NewV7(),
		OrderID:   orderID,
		ProductID: productID,
		Quantity:  quantity,
	}
}

type OrderItem struct {
	ProductID uuid.UUID
	Quantity  int
}

func (i *OrderItem) ToItem(orderID uuid.UUID) *Item {
	return NewItem(
		orderID,
		i.ProductID,
		i.Quantity,
	)
}

func ItemsFromOrderItems(orderID uuid.UUID, orderItems []OrderItem) []Item {
	return aggr.Map(orderItems, func(v OrderItem, i int) Item {
		return *v.ToItem(orderID)
	})
}
