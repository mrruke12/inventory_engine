package order

import (
	"context"
	"uuid"
)

type Repository interface {
	CreateOrder(ctx context.Context, order *Order) error
	CreateAndReserveOrderItems(ctx context.Context, products []uuid.UUID, deltas []int) error
	BatchMarkExpiredOrders(ctx context.Context, batchSize int) (int, error)
}
