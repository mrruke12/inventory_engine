//go:build integration

package infrastructure

import (
	"context"
	"testing"
	"time"
	"uuid"

	"github.com/mrruke12/inventory_engine/internal/domains/order"
	"github.com/mrruke12/inventory_engine/internal/infrastructure/postgres"
	"github.com/mrruke12/inventory_engine/test/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatchMarkExpiredOrders_MarkOnlyExpiredPendingOrders(t *testing.T) {
	integration.CleanupPostgres(t, pool)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	repo := postgres.NewOrderRepository(pool)
	var productID, pendingOrderID uuid.UUID

	err := pool.QueryRow(ctx, `
		with inserted_product as (
			insert into products (id, name, stock)
			values (uuidv7(), 'test', 0)
			returning id
		),
		inserted_orders as (
			insert into orders (id, status, expires_at)
			values 
				(uuidv7(), 'pending', now() - interval '1 hour'), -- only this order must be marked as expired
				(uuidv7(), 'pending', now() + interval '1 hour'),
				(uuidv7(), 'released', now()),
				(uuidv7(), 'released', now() + interval '1 hour'),
				(uuidv7(), 'expired', now()),
				(uuidv7(), 'expired', now() + interval '1 hour'),
				(uuidv7(), 'completed', now()),
				(uuidv7(), 'completed', now() + interval '1 hour'),
				(uuidv7(), 'canceled', now()),
				(uuidv7(), 'canceled', now() + interval '1 hour')
			returning id, status, expires_at
		),
		inserted_items as (
			insert into order_items (id, order_id, product_id, amount)
			select 
				o.id as id,
				o.id as order_id,
				(select id from inserted_product) as product_id,
				(1) as amount
			from inserted_orders o
		)
		select
			(select id from inserted_product) as product_id,
			(
				select id from inserted_orders
				where status = 'pending' and expires_at < now()
			) as order_id
	`).Scan(&productID, &pendingOrderID)

	require.NoError(t, err, "insert orders")

	updated, err := repo.BatchMarkExpiredOrders(ctx, 10)

	require.NoError(t, err)
	assert.Equal(t, 1, updated, "updated orders")

	var status order.Status
	var stock int

	err = pool.QueryRow(ctx, `
		select
			(
				select status from orders
				where id = $1
			) as status,
			(
				select stock from products
				where id = $2
			) as stock
	`, pendingOrderID, productID).Scan(&status, &stock)

	require.NoError(t, err, "query results")

	assert.Equal(t, order.StatusExpired, status, "order status")
	assert.Equal(t, 1, stock, "stock")
}

func TestBatchMarkExpiredOrders_MarkAtMostBatchSizeOrders(t *testing.T) {
	integration.CleanupPostgres(t, pool)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	repo := postgres.NewOrderRepository(pool)
	var productID uuid.UUID
	batchSize := 3

	err := pool.QueryRow(ctx, `
		with inserted_product as (
			insert into products (id, name, stock)
			values (uuidv7(), 'test', 0)
			returning id
		),
		inserted_orders as (
			insert into orders (id, status, expires_at)
			values 
				(uuidv7(), 'pending', now() - interval '1 hour'),
				(uuidv7(), 'pending', now() - interval '1 hour'),
				(uuidv7(), 'pending', now() - interval '1 hour'),
				(uuidv7(), 'pending', now() - interval '1 hour'),
				(uuidv7(), 'pending', now() - interval '1 hour')
			returning id, status, expires_at
		),
		inserted_items as (
			insert into order_items (id, order_id, product_id, amount)
			select 
				o.id as id,
				o.id as order_id,
				(select id from inserted_product) as product_id,
				(1) as amount
			from inserted_orders o
		)
		select id as product_id from inserted_product
	`).Scan(&productID)

	require.NoError(t, err, "insert orders")

	updated, err := repo.BatchMarkExpiredOrders(ctx, batchSize)

	require.NoError(t, err)
	assert.Equal(t, batchSize, updated, "updated orders")

	var expiredOrders, stock int

	err = pool.QueryRow(ctx, `
		select
			(
				select count(*) from orders
				where status = 'expired'
			) as expiredOrders,
			(
				select stock from products
				where id = $1
			) as stock
	`, productID).Scan(&expiredOrders, &stock)

	require.NoError(t, err, "query results")

	assert.Equal(t, batchSize, expiredOrders, "expired orders")
	assert.Equal(t, batchSize, stock, "stock")
}

func TestBatchMarkExpiredOrders_SkipLockedOrders(t *testing.T) {
	integration.CleanupPostgres(t, pool)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	var orderID uuid.UUID

	err := pool.QueryRow(ctx, `
		with inserted_product as (
			insert into products (id, name, stock)
			values (uuidv7(), 'test', 0)
			returning id
		),
		inserted_orders as (
			insert into orders (id, status, expires_at)
			values 
				(uuidv7(), 'pending', now() - interval '1 hour')
			returning id, status, expires_at
		),
		inserted_items as (
			insert into order_items (id, order_id, product_id, amount)
			select 
				o.id as id,
				o.id as order_id,
				(select id from inserted_product) as product_id,
				(1) as amount
			from inserted_orders o
		)
		select 
			(select id from inserted_orders) as order_id

	`).Scan(&orderID)

	require.NoError(t, err, "insert orders")

	repo := postgres.NewOrderRepository(pool)

	func() {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)

		defer tx.Rollback(ctx)

		_, err = tx.Exec(ctx, `select * from orders where id = $1 for update`, orderID)
		require.NoError(t, err)

		updated, err := repo.BatchMarkExpiredOrders(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, 0, updated)

		tx.Commit(ctx)
	}()
}
