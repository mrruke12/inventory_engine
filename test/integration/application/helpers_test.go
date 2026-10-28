//go:build integration

package application

import (
	"context"
	"testing"
	"uuid"

	"github.com/mrruke12/inventory_engine/internal/application/orderusecase"
	"github.com/mrruke12/inventory_engine/internal/domains/order"
	"github.com/stretchr/testify/require"
)

func requireTestProduct(ctx context.Context, t *testing.T, stock int) (productID uuid.UUID) {
	productID = uuid.NewV7()

	sql := `
		insert into products (id, name, stock)
		values ($1, $2, $3)
	`

	_, err := pool.Exec(ctx, sql, productID, productID.String(), stock)

	require.NoError(t, err, "could not insert a test product")

	return productID
}

func requireProductStockAndOrders(ctx context.Context, t *testing.T, productID uuid.UUID) (stock, orders int) {
	sql := `
		select 
			(select stock from products where id = $1) as stock,
			(
				select count(o.*) from orders o 
				where exists (
					select 1
					from order_items i
					where i.order_id = o.id
						and i.product_id = $1
				)
			) as orders
	`

	err := pool.QueryRow(ctx, sql, productID).Scan(&stock, &orders)

	require.NoError(t, err, "could not get result stock")

	return stock, orders
}

func createOrder(ctx context.Context, productID uuid.UUID, idempotencyKey string, amount int) (uuid.UUID, error) {
	return orderService.CreateOrder(ctx, orderusecase.CreateOrderCommand{
		IdempotencyKey: idempotencyKey,
		OrderItems: []order.OrderItem{{
			ProductID: productID,
			Amount:    amount,
		}},
	})
}
