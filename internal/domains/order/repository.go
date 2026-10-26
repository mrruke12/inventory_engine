package order

import (
	"context"
	"uuid"
)

type Repository interface {
	CreateOrder(ctx context.Context, order *Order) error
	CreateAndReserveOrderItems(ctx context.Context, items []Item) error
	BatchCancelExpiredOrders(ctx context.Context, batchSize int) ([]int, error)

	// Releases an order if it's still pending; no-op otherwise. Returns:
	//  1. initialStatus - initial order status; empty string if err is not nil
	//  2. canceled - is the status changed; false if err is not nil
	//  3. err - driver error
	ReleaseOrder(ctx context.Context, orderID uuid.UUID) (initialStatus Status, updated bool, err error)
}
