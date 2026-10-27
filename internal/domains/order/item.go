package order

import (
	"uuid"

	"github.com/mrruke12/inventory_engine/pkg/aggr"
)

type Item struct {
	ID        uuid.UUID
	OrderID   uuid.UUID
	ProductID uuid.UUID
	Amount    int
}

func NewItem(orderID, productID uuid.UUID, amount int) *Item {
	return &Item{
		ID:        uuid.NewV7(),
		OrderID:   orderID,
		ProductID: productID,
		Amount:    amount,
	}
}

type OrderItem struct {
	ProductID uuid.UUID
	Amount    int
}

func (i *OrderItem) ToItem(orderID uuid.UUID) *Item {
	return NewItem(
		orderID,
		i.ProductID,
		i.Amount,
	)
}

func ItemsFromOrderItems(orderID uuid.UUID, orderItems []OrderItem) []Item {
	return aggr.Map(orderItems, func(v OrderItem, i int) Item {
		return *v.ToItem(orderID)
	})
}
