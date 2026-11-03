package mocks

import (
	"context"
	"uuid"

	"github.com/mrruke12/inventory_engine/internal/domains/order"
)

type OrderRepository struct {
	CreateOrderFunc                func(ctx context.Context, order *order.Order) error
	CreateAndReserveOrderItemsFunc func(ctx context.Context, orderID uuid.UUID, items []order.Item) error
	BatchMarkExpiredOrdersFunc     func(ctx context.Context, batchSize int) (updatedOrders int, err error)
	ReleaseOrderFunc               func(ctx context.Context, orderID uuid.UUID) (result order.PendingOrderUpdateResult, err error)
	CompleteOrderFunc              func(ctx context.Context, orderID uuid.UUID) (result order.PendingOrderUpdateResult, err error)
	CancelOrderFunc                func(ctx context.Context, orderID uuid.UUID) (result order.CancelOrderResult, err error)
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *order.Order) error {
	if r.CreateOrderFunc == nil {
		panic("CreateOrderFunc is not set")
	}
	return r.CreateOrderFunc(ctx, order)
}

func (r *OrderRepository) CreateAndReserveOrderItems(ctx context.Context, orderID uuid.UUID, items []order.Item) error {
	if r.CreateAndReserveOrderItemsFunc == nil {
		panic("CreateAndReserveOrderItemsFunc is not set")
	}
	return r.CreateAndReserveOrderItemsFunc(ctx, orderID, items)
}

func (r *OrderRepository) BatchMarkExpiredOrders(ctx context.Context, batchSize int) (updatedOrders int, err error) {
	if r.BatchMarkExpiredOrdersFunc == nil {
		panic("BatchMarkExpiredOrdersFunc is not set")
	}
	return r.BatchMarkExpiredOrdersFunc(ctx, batchSize)
}

func (r *OrderRepository) ReleaseOrder(ctx context.Context, orderID uuid.UUID) (result order.PendingOrderUpdateResult, err error) {
	if r.ReleaseOrderFunc == nil {
		panic("ReleaseOrderFunc is not set")
	}
	return r.ReleaseOrderFunc(ctx, orderID)
}

func (r *OrderRepository) CompleteOrder(ctx context.Context, orderID uuid.UUID) (result order.PendingOrderUpdateResult, err error) {
	if r.CompleteOrderFunc == nil {
		panic("CompleteOrderFunc is not set")
	}
	return r.CompleteOrderFunc(ctx, orderID)
}

func (r *OrderRepository) CancelOrder(ctx context.Context, orderID uuid.UUID) (result order.CancelOrderResult, err error) {
	if r.CancelOrderFunc == nil {
		panic("CancelOrderFunc is not set")
	}
	return r.CancelOrderFunc(ctx, orderID)
}
