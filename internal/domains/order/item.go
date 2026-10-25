package order

import (
	"uuid"

	"github.com/mrruke12/inventory_engine/pkg/aggr"
)

type Item struct {
	ID        uuid.UUID
	OrderID   string
	ProductID uuid.UUID
	Quantity  int
}

func NewItem(orderID string, productID uuid.UUID, quantity int) *Item {
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

func (i *OrderItem) ToItem(orderID string) *Item {
	return NewItem(
		orderID,
		i.ProductID,
		i.Quantity,
	)
}

func ItemsFromOrderItems(orderID string, orderItems []OrderItem) []Item {
	return aggr.Map(orderItems, func(v OrderItem, i int) Item {
		return *v.ToItem(orderID)
	})
}
