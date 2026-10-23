package order

import (
	"context"
)

type Repository interface {
	CreateOrder(ctx context.Context, order *Order) error
	CreateAndReserveOrderItems(ctx context.Context, items []Item) error
	BatchMarkExpiredOrders(ctx context.Context, batchSize int) (int, error)
}
