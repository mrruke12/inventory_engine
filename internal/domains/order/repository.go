package order

import (
	"context"
	"uuid"
)

type PendingOrderUpdateResult struct {
	InitialStatus *Status
	Expired       bool
	Updated       bool
	Found         bool
}

type CancelOrderResult struct {
	InitialStatus *Status
	Updated       bool
	Found         bool
}

type Repository interface {
	CreateOrder(ctx context.Context, order *Order) error
	CreateAndReserveOrderItems(ctx context.Context, orderID uuid.UUID, items []Item) error
	BatchMarkExpiredOrders(ctx context.Context, batchSize int) (updatedOrders int, err error)
	ReleaseOrder(ctx context.Context, orderID uuid.UUID) (result PendingOrderUpdateResult, err error)
	CompleteOrder(ctx context.Context, orderID uuid.UUID) (result PendingOrderUpdateResult, err error)
	CancelOrder(ctx context.Context, orderID uuid.UUID) (result CancelOrderResult, err error)
}
