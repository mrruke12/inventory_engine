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
	//  1. initialStatus - initial order status (nil if found is false or error is not nil);
	//  1. updated - is the status changed (false if found is false or err is not nil);
	//  1. found - is the order found;
	//  1. err - driver error.
	ReleaseOrder(ctx context.Context, orderID uuid.UUID) (initialStatus *Status, updated bool, found bool, err error)
}
